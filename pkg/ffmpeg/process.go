package ffmpeg

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

// ProcessManager 管理 FFmpeg 进程，防止僵尸进程
type ProcessManager struct {
	processes map[int]*ManagedProcess
	mutex     sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
}

// ManagedProcess 被管理的进程
type ManagedProcess struct {
	cmd       *exec.Cmd
	pid       int
	startTime time.Time
	ctx       context.Context
	cancel    context.CancelFunc
	done      chan error
	cleanup   func()
}

// NewProcessManager 创建新的进程管理器
func NewProcessManager() *ProcessManager {
	ctx, cancel := context.WithCancel(context.Background())
	pm := &ProcessManager{
		processes: make(map[int]*ManagedProcess),
		ctx:       ctx,
		cancel:    cancel,
	}

	// 启动清理协程
	go pm.cleanupRoutine()

	return pm
}

// StartProcess 启动一个受管理的 FFmpeg 进程
func (pm *ProcessManager) StartProcess(ctx context.Context, name string, args []string, env []string) (*ManagedProcess, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = append(os.Environ(), env...)

	// 设置进程组，便于管理
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	// 创建进程上下文
	procCtx, cancel := context.WithCancel(ctx)

	mp := &ManagedProcess{
		cmd:       cmd,
		startTime: time.Now(),
		ctx:       procCtx,
		cancel:    cancel,
		done:      make(chan error, 1),
	}

	// 启动进程
	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("启动进程失败: %w", err)
	}

	mp.pid = cmd.Process.Pid

	// 注册进程
	pm.mutex.Lock()
	pm.processes[mp.pid] = mp
	pm.mutex.Unlock()

	// 监控进程结束
	go func() {
		err := cmd.Wait()
		mp.done <- err

		// 从管理器中移除
		pm.mutex.Lock()
		delete(pm.processes, mp.pid)
		pm.mutex.Unlock()

		// 执行清理
		if mp.cleanup != nil {
			mp.cleanup()
		}
	}()

	return mp, nil
}

// TerminateProcess 终止进程
func (pm *ProcessManager) TerminateProcess(pid int) error {
	pm.mutex.RLock()
	mp, exists := pm.processes[pid]
	pm.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("进程 %d 不存在", pid)
	}

	// 取消上下文
	mp.cancel()

	// 发送 SIGTERM 到进程组
	if mp.cmd.Process != nil {
		pgid, err := syscall.Getpgid(pid)
		if err == nil {
			syscall.Kill(-pgid, syscall.SIGTERM)
		}
	}

	// 等待进程结束或超时
	select {
	case <-mp.done:
		return nil
	case <-time.After(5 * time.Second):
		// 强制杀死进程
		if mp.cmd.Process != nil {
			pgid, err := syscall.Getpgid(pid)
			if err == nil {
				syscall.Kill(-pgid, syscall.SIGKILL)
			}
		}
		return nil
	}
}

// KillAllProcesses 杀死所有管理的进程
func (pm *ProcessManager) KillAllProcesses() {
	pm.mutex.RLock()
	pids := make([]int, 0, len(pm.processes))
	for pid := range pm.processes {
		pids = append(pids, pid)
	}
	pm.mutex.RUnlock()

	for _, pid := range pids {
		pm.TerminateProcess(pid)
	}
}

// GetProcessCount 获取当前管理的进程数量
func (pm *ProcessManager) GetProcessCount() int {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	return len(pm.processes)
}

// cleanupRoutine 定期清理僵尸进程
func (pm *ProcessManager) cleanupRoutine() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-pm.ctx.Done():
			return
		case <-ticker.C:
			pm.cleanupZombies()
		}
	}
}

// cleanupZombies 清理僵尸进程
func (pm *ProcessManager) cleanupZombies() {
	pm.mutex.RLock()
	processes := make([]*ManagedProcess, 0, len(pm.processes))
	for _, mp := range pm.processes {
		processes = append(processes, mp)
	}
	pm.mutex.RUnlock()

	for _, mp := range processes {
		select {
		case err := <-mp.done:
			if err != nil {
				fmt.Printf("进程 %d 异常退出: %v\n", mp.pid, err)
			}
		default:
			// 进程仍在运行
		}
	}
}

// Close 关闭进程管理器
func (pm *ProcessManager) Close() error {
	pm.cancel()
	pm.KillAllProcesses()
	return nil
}

// ManagedProcess 方法

// PID 返回进程 ID
func (mp *ManagedProcess) PID() int {
	return mp.pid
}

// Wait 等待进程结束
func (mp *ManagedProcess) Wait() error {
	return <-mp.done
}

// Terminate 终止进程
func (mp *ManagedProcess) Terminate() error {
	mp.cancel()
	return nil
}

// IsRunning 检查进程是否在运行
func (mp *ManagedProcess) IsRunning() bool {
	select {
	case <-mp.done:
		return false
	default:
		return true
	}
}

// SetCleanup 设置清理函数
func (mp *ManagedProcess) SetCleanup(cleanup func()) {
	mp.cleanup = cleanup
}
