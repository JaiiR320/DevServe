package internal

var LockToPM = map[string]string{
	"package-lock.json": "npm",
	"pnpm-lock.yaml":    "pnpm",
	"yarn.lock":         "yarn",
	"bun.lock":          "bun",
	"bun.lockb":         "bun",
}

var PMToCommand = map[string]string{
	"npm":  "npm run dev",
	"pnpm": "pnpm run dev",
	"yarn": "yarn dev",
	"bun":  "bun run dev",
}
