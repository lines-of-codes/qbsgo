# An overcomplicated build script

import os
import shutil

version = "1.0.0"
systems = {
    "windows": ["amd64", "arm64"],
    "darwin": ["amd64", "arm64"],
    "linux": ["amd64", "arm64", "arm"],
}

shutil.rmtree("./dist")

# "system" = Operating System
for system, arches in systems.items():
    fileExt = ".exe" if system == "windows" else ""

    for arch in arches:
        outDir = f"./dist/qbsgo_{system}_{arch}/"

        os.makedirs(outDir, exist_ok=True)

        cmd = f"GOOS={system} GOARCH={arch} go build -o {outDir}qbsgo{fileExt}"

        print("Running: " + cmd)

        print(f"Return Value: {os.system(cmd)}")

        _ = shutil.copy2("LICENSE.md", outDir)

        archiveCmd = f"tar -I 'zstd --ultra -22' -cf ./dist/qbsgo_{version}_{system}_{arch}.tar.zst --directory=./dist qbsgo_{system}_{arch}/"

        # Only include example configuration for Linux because the paths inside
        # is configured for Linux
        if system == "linux":
            _ = shutil.copy2("qbsgo.example.toml", outDir + "qbsgo.toml")

        if system == "windows":
            archiveCmd = (
                f"zip -9 -r ./dist/qbsgo_{version}_{system}_{arch}.zip {outDir}"
            )

        print("Running: " + archiveCmd)
        print(f"Archive creation returned: {os.system(archiveCmd)}")
