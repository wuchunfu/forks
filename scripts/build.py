#!/usr/bin/env python3
"""Forks 构建脚本 — 多平台交叉编译，自动注入版本号

用法:
    python scripts/build.py                     # 构建所有平台
    python scripts/build.py --local             # 仅构建当前平台
    python scripts/build.py --platform windows  # 仅构建 Windows
    python scripts/build.py --skip-frontend     # 跳过前端构建
"""

import os
import shutil
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
DIST_DIR = ROOT / "dist"

TARGETS = [
    ("windows", "amd64"),
    ("linux", "amd64"),
    ("linux", "arm64"),
    ("darwin", "amd64"),
    ("darwin", "arm64"),
]


def read_version():
    v = (ROOT / "VERSION").read_text().strip()
    return v if v else "dev"


def get_commit():
    ret = subprocess.run(["git", "rev-parse", "--short", "HEAD"],
                         capture_output=True, text=True, cwd=ROOT)
    return ret.stdout.strip() if ret.returncode == 0 else "unknown"


def build_frontend():
    dist = ROOT / "web" / "dist"
    if dist.is_dir() and any(dist.iterdir()):
        print("前端已构建，跳过")
        return
    print(">>> 构建前端...")
    subprocess.run(["npm", "run", "build"], cwd=str(ROOT / "web"), check=True)


def build_target(goos, goarch, version, commit):
    ext = ".exe" if goos == "windows" else ""
    output_name = f"forks_{goos}_{goarch}{ext}"
    output_path = DIST_DIR / output_name

    ldflags = (
        f"-s -w "
        f"-X github.com/cicbyte/forks/utils.Version={version}"
    )

    env = {**os.environ, "GOOS": goos, "GOARCH": goarch, "CGO_ENABLED": "0"}

    print(f"  编译 {goos}/{goarch} ...", end=" ", flush=True)
    ret = subprocess.run(
        ["go", "build", f"-ldflags={ldflags}", "-o", str(output_path), "."],
        cwd=str(ROOT), env=env
    )
    if ret.returncode != 0:
        print(f"失败!")
        return False

    size_mb = output_path.stat().st_size / 1024 / 1024
    print(f"OK ({size_mb:.1f} MB) -> dist/{output_name}")
    return True


def main():
    import argparse
    parser = argparse.ArgumentParser(description="Forks 构建脚本")
    parser.add_argument("--platform", choices=["windows", "linux", "darwin"], help="仅编译指定平台")
    parser.add_argument("--local", action="store_true", help="仅编译当前平台")
    parser.add_argument("--skip-frontend", action="store_true", help="跳过前端构建")
    args = parser.parse_args()

    os.chdir(str(ROOT))

    version = read_version()
    commit = get_commit()
    print(f"forks {version} | commit: {commit}")
    print()

    if not args.skip_frontend:
        build_frontend()
        print()

    if DIST_DIR.exists():
        shutil.rmtree(DIST_DIR)
    DIST_DIR.mkdir()

    targets = TARGETS
    if args.local:
        current = "windows" if os.name == "nt" else ("linux" if os.name == "posix" else "darwin")
        targets = [(current, "amd64")]
    elif args.platform:
        targets = [(args.platform, arch) for p, arch in TARGETS if p == args.platform]

    success = True
    for goos, goarch in targets:
        if not build_target(goos, goarch, version, commit):
            success = False

    print()
    if success:
        print(f"构建完成! 输出目录: dist/")
    else:
        print("部分平台构建失败!")
        sys.exit(1)


if __name__ == "__main__":
    main()
