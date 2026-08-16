# Repository Notes

## Windows runtime layout

- Scoop/original GoSuki uses `D:\C2D\dotfiles\gosuki\config.toml`, `%APPDATA%\gosuki\gosuki.db`, and `127.0.0.1:2025`.
- Fork/debug GoSuki uses `D:\C2D\dotfiles\gosuki-dev`, including its own database, imports, logs, and `127.0.0.1:22025`.
- Debug launch scripts must validate the `gosuki-fork` role marker and must never stop or replace the Scoop executable.

## Windows debug workflow

- Build with `./build-debug.ps1`.
- Start the isolated executable with `./start-debug.ps1`.
- Stop only that executable with `./stop-debug.ps1`.
- Rebuild and restart with `./restart-debug.ps1`.
- The debug binary is `build\gosuki-debug.exe` and is intentionally ignored by Git.
