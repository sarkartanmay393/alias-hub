package server

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/bash-shortcuts/ah/pkg/manager"
	"github.com/bash-shortcuts/ah/pkg/parser"
)

//go:embed web_dist/*
var webFS embed.FS

// NOTE: Since I cannot easily structure the previously created `web/` folder to match `embed`,
// I will assume the build process or I just point `webFS` to `../../web`.
// Actually, `embed` directives are relative to the file.
// If I put this file in `pkg/server`, and assets are in `root/web`.
// Go generic embed doesn't support '..'.
// Optimization: I should move `web` contents to `pkg/server/web` OR
// I will move `web/` to `pkg/server/web_dist` during the "Build" phase?
// EASIER: I'll create `pkg/server/web_assets.go` and COPY the web content there OR
// Just move the `web` folder inside `pkg/server`.
// I'll stick to: Move `web` to `pkg/server/web` in the next step.

type Conflict struct {
	Alias    string  `json:"alias"`
	Existing PkgInfo `json:"existing"`
	New      PkgInfo `json:"new"`
}

type PkgInfo struct {
	Package string `json:"package"`
	Command string `json:"command"`
}

type ResolveRequest struct {
	Alias         string `json:"alias"`
	Action        string `json:"action"` // "keep_existing", "replace", "rename:<new_name>"
	TargetPackage string `json:"targetPackage"`
}

var currentConflicts []Conflict

// Start launches the conflict resolution server
func Start(newPkgName string) error {
	// 1. Calculate Conflicts
	var err error
	currentConflicts, err = calculateConflicts(newPkgName)
	if err != nil {
		return err
	}

	if len(currentConflicts) == 0 {
		fmt.Println("No conflicts found!")
		return nil
	}

	// 2. Setup Server
	// We need to serve the "web" folder.
	// For now, let's assume assets are in a subfolder "web" relative to this binary execution or embedded.
	// Since I can't move files easily with `embed` restrictions in this chat mode without multiple steps,
	// I'll assume they will be moved.

	http.Handle("/", http.FileServer(http.FS(getWebFS())))
	http.HandleFunc("/api/conflicts", handleConflicts)
	http.HandleFunc("/api/shutdown", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		go func() {
			fmt.Println("Shutting down...")
			os.Exit(0)
		}()
	})

	port := "9999"
	url := "http://localhost:" + port
	fmt.Printf("⚠️  Conflict Resolution UI started at %s\n", url)
	openBrowser(url)

	// Security: Only bind to localhost to prevent network access
	return http.ListenAndServe("127.0.0.1:"+port, nil)
}

func handleConflicts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(currentConflicts)
}

func handleResolve(w http.ResponseWriter, r *http.Request) {
	var req ResolveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	fmt.Printf("Action Received: %s for %s (Target: %s)\n", req.Action, req.Alias, req.TargetPackage)

	// Resolution Logic
	switch req.Action {
	case "replace":
		// User chose the NEW package.
		// We need to Enable the new package.
		// NOTE: This might cause OTHER conflicts if the new package has other aliases.
		// But for the specific conflict at hand, this resolves it.
		// We use the Atomic EnablePackage.
		if err := manager.EnablePackage(req.TargetPackage); err != nil {
			http.Error(w, fmt.Sprintf("Failed to enable package: %v", err), 500)
			return
		}
		fmt.Printf("✅ Resolved: Switched to package '%s'\n", req.TargetPackage)

	case "keep_existing":
		// User chose to KEEP valid installation.
		// We don't need to do anything technically, as the existing one is already active.
		// But we might want to "mark" it resolved in the GUI?
		// The GUI removes it from the list client-side.
		fmt.Printf("✅ Resolved: Kept existing package for '%s'\n", req.Alias)

	case "rename":
		// Not supported yet in this architecture.
		http.Error(w, "Rename not supported in this version", 400)
		return

	default:
		http.Error(w, "Unknown action", 400)
		return
	}

	w.WriteHeader(200)
}

func calculateConflicts(pkgName string) ([]Conflict, error) {
	// Logic to actually parse active packages vs new package
	// Re-using manager.CheckConflicts logic but constructing detailed objects
	// For this demo/first-pass, I'll return dummy data if pkgName is "demo",
	// or try to run real check.

	registryPath, err := manager.GetRegistryPackagePath(pkgName)
	if err != nil {
		return nil, err
	}

	rawConflicts, err := manager.CheckConflicts(registryPath)
	if err != nil {
		return nil, err
	}

	var list []Conflict
	newAliases, _ := parser.ParseAliases(filepath.Join(registryPath, "alias.sh"))

	// We need to find WHAT the existing alias command is.
	// manager.CheckConflicts only returns "pkgName".
	// We need to scan again.

	for alias, existingPkg := range rawConflicts {
		// Find New Command
		var newCmd string
		for _, a := range newAliases {
			if a.Name == alias {
				newCmd = a.Command
				break
			}
		}

		// Find Existing Command
		root, _ := manager.GetRootDir()
		existingPath := filepath.Join(root, "active", existingPkg, "alias.sh")
		existParams, _ := parser.ParseAliases(existingPath)
		var existCmd string
		for _, a := range existParams {
			if a.Name == alias {
				existCmd = a.Command
				break
			}
		}

		list = append(list, Conflict{
			Alias:    alias,
			Existing: PkgInfo{Package: existingPkg, Command: existCmd},
			New:      PkgInfo{Package: pkgName, Command: newCmd},
		})
	}

	return list, nil
}

func getWebFS() fs.FS {
	// Hack: To make it work in this environment without moving files yet
	// In production this would return the embedded FS
	f, err := fs.Sub(webFS, "web_dist")
	if err == nil {
		return f
	}
	// Fallback to local dir
	return os.DirFS("web")
}

func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	}
	if err != nil {
		fmt.Println("Error opening browser:", err)
	}
}
