package commands


// SkipDirs contains all directory names that should be ignored during traversal.
var SkipDirs = map[string]struct{}{
	".git": {}, "node_modules": {}, "vendor": {}, "dist": {}, "build": {},
	"out": {}, "target": {}, "bin": {}, "obj": {},
	".next": {}, ".nuxt": {}, ".output": {}, ".svelte-kit": {},
	".vite": {}, ".turbo": {}, ".cache": {}, ".parcel-cache": {},
	".swc": {}, ".eslintcache": {}, ".cogito": {},

	// Python
	"__pycache__": {}, ".venv": {}, "venv": {}, "env": {},
	".pytest_cache": {}, ".mypy_cache": {}, ".ruff_cache": {}, ".ipynb_checkpoints": {},

	// JS / TS
	".tsbuildinfo": {}, ".webpack": {}, ".rollup.cache": {},
	".babel-cache": {}, ".storybook": {}, ".angular": {},
	".expo": {}, ".expo-shared": {},

	// IDE
	".vscode": {}, ".idea": {}, ".fleet": {}, ".vs": {},
	".sublime-project": {}, ".sublime-workspace": {},

	// testing / logs
	"coverage": {}, ".nyc_output": {}, "logs": {},
	"log": {}, "tmp": {}, "temp": {},

	// OS junk
	".DS_Store": {}, "Thumbs.db": {}, "desktop.ini": {},

	// DevOps
	".docker": {}, ".terraform": {}, ".serverless": {},
	".pulumi": {}, ".kube": {}, ".aws-sam": {},

	// build tools
	".gradle": {}, ".maven": {}, ".cargo": {},
	".go-build": {}, ".zig-cache": {},
}

// ShouldSkipDir checks if a directory should be skipped.
func ShouldSkipDir(name string) bool {
	_, ok := SkipDirs[name]
	return ok
}