package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	version   = "1.0.0"
	signature = "Designed by Ifesinachi Osude"
)

type UpdateMetrics struct {
	UpdatesAvailable    int
	PendingCount        int
	KernelAvailable     int
	CurrentKernelVer    string
	PendingKernelVer    string
	SecurityAvailable   int
	SecurityCount       int
	RebootRequired      int
	CheckError          int
	PendingPackages     []PackageInfo
}

type PackageInfo struct {
	NameArch string
	Name     string
	Arch     string
	Version  string
	Repo     string
}

func escapeLabel(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
}

func getCurrentKernelVersion() string {
	cmd := exec.Command("uname", "-r")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func checkUpdates() (*UpdateMetrics, error) {
	metrics := &UpdateMetrics{}
	
	// Detect package manager
	pkgMgr := "dnf"
	if _, err := exec.LookPath("dnf"); err != nil {
		pkgMgr = "yum"
	}

	// Get current kernel
	metrics.CurrentKernelVer = getCurrentKernelVersion()

	// Run check-update
	cmd := exec.Command(pkgMgr, "-q", "check-update")
	output, err := cmd.CombinedOutput()
	
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			metrics.CheckError = 1
			return metrics, nil
		}
	}

	// Exit code 100 means updates available
	if exitCode == 100 {
		metrics.UpdatesAvailable = 1
	} else if exitCode != 0 {
		metrics.CheckError = 1
		return metrics, nil
	}

	// Parse output
	if metrics.UpdatesAvailable == 1 {
		scanner := bufio.NewScanner(strings.NewReader(string(output)))
		for scanner.Scan() {
			line := scanner.Text()
			
			// Skip empty lines
			if line == "" {
				continue
			}
			
			// Stop at Obsoleting section
			if strings.HasPrefix(line, "Obsoleting") {
				break
			}
			
			fields := strings.Fields(line)
			if len(fields) < 3 {
				continue
			}
			
			// Must have dot in first field (name.arch)
			if !strings.Contains(fields[0], ".") {
				continue
			}

			pkg := PackageInfo{
				NameArch: fields[0],
				Version:  fields[1],
			}
			
			if len(fields) >= 3 {
				pkg.Repo = fields[2]
			}

			// Split name.arch
			parts := strings.Split(pkg.NameArch, ".")
			if len(parts) >= 2 {
				pkg.Arch = parts[len(parts)-1]
				pkg.Name = strings.Join(parts[:len(parts)-1], ".")
			}

			metrics.PendingPackages = append(metrics.PendingPackages, pkg)
			metrics.PendingCount++

			// Check for kernel updates (prefer kernel-core, then kernel)
			if pkg.Name == "kernel-core" && metrics.PendingKernelVer == "" {
				metrics.PendingKernelVer = pkg.Version
				metrics.KernelAvailable = 1
			} else if pkg.Name == "kernel" && metrics.PendingKernelVer == "" {
				metrics.PendingKernelVer = pkg.Version
				metrics.KernelAvailable = 1
			}
		}
	}

	// Check security updates (dnf only)
	if pkgMgr == "dnf" {
		cmd := exec.Command("dnf", "-q", "updateinfo", "list", "security")
		output, err := cmd.Output()
		if err == nil && len(output) > 0 {
			count := 0
			scanner := bufio.NewScanner(strings.NewReader(string(output)))
			for scanner.Scan() {
				if strings.TrimSpace(scanner.Text()) != "" {
					count++
				}
			}
			metrics.SecurityCount = count
			if count > 0 {
				metrics.SecurityAvailable = 1
			}
		}
	}

	// Check reboot required
	if _, err := exec.LookPath("needs-restarting"); err == nil {
		cmd := exec.Command("needs-restarting", "-r")
		err := cmd.Run()
		if err != nil {
			metrics.RebootRequired = 1
		}
	} else {
		// Fallback: compare running kernel to newest installed
		cmd := exec.Command("rpm", "-q", "--last", "kernel-core")
		output, err := cmd.Output()
		if err == nil && len(output) > 0 {
			scanner := bufio.NewScanner(strings.NewReader(string(output)))
			if scanner.Scan() {
				line := scanner.Text()
				fields := strings.Fields(line)
				if len(fields) > 0 {
					newestKernel := strings.TrimPrefix(fields[0], "kernel-core-")
					if metrics.CurrentKernelVer != "" && !strings.HasPrefix(metrics.CurrentKernelVer, newestKernel) {
						metrics.RebootRequired = 1
					}
				}
			}
		}
	}

	return metrics, nil
}

func main() {
	metrics, err := checkUpdates()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		metrics = &UpdateMetrics{CheckError: 1}
	}

	// Output Prometheus metrics
	fmt.Println("# TYPE dnf_update_available gauge")
	fmt.Printf("dnf_update_available %d\n", metrics.UpdatesAvailable)

	fmt.Println("# TYPE dnf_update_pending_count gauge")
	fmt.Printf("dnf_update_pending_count %d\n", metrics.PendingCount)

	fmt.Println("# TYPE dnf_update_kernel_available gauge")
	fmt.Printf("dnf_update_kernel_available %d\n", metrics.KernelAvailable)

	// Kernel version info (matching shell script output)
	fmt.Println("# TYPE dnf_update_kernel_version_info gauge")
	if metrics.KernelAvailable == 1 {
		pendingKernelEsc := escapeLabel(metrics.PendingKernelVer)
		fmt.Printf("dnf_update_kernel_version_info{version=\"%s\"} 1\n", pendingKernelEsc)
	} else {
		fmt.Printf("dnf_update_kernel_version_info{version=\"\"} 0\n")
	}

	// Kernel version info - BOTH current and pending (additional metrics)
	fmt.Println("# TYPE dnf_update_kernel_current_version_info gauge")
	currentKernelEsc := escapeLabel(metrics.CurrentKernelVer)
	fmt.Printf("dnf_update_kernel_current_version_info{version=\"%s\"} 1\n", currentKernelEsc)

	fmt.Println("# TYPE dnf_update_kernel_pending_version_info gauge")
	if metrics.KernelAvailable == 1 {
		pendingKernelEsc := escapeLabel(metrics.PendingKernelVer)
		fmt.Printf("dnf_update_kernel_pending_version_info{version=\"%s\"} 1\n", pendingKernelEsc)
	} else {
		fmt.Printf("dnf_update_kernel_pending_version_info{version=\"\"} 0\n")
	}

	fmt.Println("# TYPE dnf_update_security_available gauge")
	fmt.Printf("dnf_update_security_available %d\n", metrics.SecurityAvailable)

	fmt.Println("# TYPE dnf_update_security_count gauge")
	fmt.Printf("dnf_update_security_count %d\n", metrics.SecurityCount)

	fmt.Println("# TYPE dnf_update_reboot_required gauge")
	fmt.Printf("dnf_update_reboot_required %d\n", metrics.RebootRequired)

	fmt.Println("# TYPE dnf_update_check_error gauge")
	fmt.Printf("dnf_update_check_error %d\n", metrics.CheckError)

	// Per-package metrics
	fmt.Println("# TYPE dnf_update_pending_pkg gauge")
	for _, pkg := range metrics.PendingPackages {
		nameEsc := escapeLabel(pkg.Name)
		archEsc := escapeLabel(pkg.Arch)
		verEsc := escapeLabel(pkg.Version)
		repoEsc := escapeLabel(pkg.Repo)
		
		fmt.Printf("dnf_update_pending_pkg{name=\"%s\",arch=\"%s\",version=\"%s\",repo=\"%s\"} 1\n",
			nameEsc, archEsc, verEsc, repoEsc)
	}

	// Truncated package list
	pkgList := ""
	for i, pkg := range metrics.PendingPackages {
		if i > 0 {
			pkgList += ","
		}
		pkgList += fmt.Sprintf("%s-%s", pkg.NameArch, pkg.Version)
		if len(pkgList) > 1200 {
			break
		}
	}
	pkgListEsc := escapeLabel(pkgList)
	fmt.Println("# TYPE dnf_update_pending_info gauge")
	fmt.Printf("dnf_update_pending_info{packages=\"%s\"} 1\n", pkgListEsc)

	fmt.Printf("# VERSION %s\n", version)
	fmt.Printf("# %s\n", signature)
}
