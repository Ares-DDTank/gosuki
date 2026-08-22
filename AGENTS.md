# Repository Notes

## Windows runtime layout

- Scoop/original GoSuki uses `D:\C2D\dotfiles\gosuki\config.toml`, `%APPDATA%\gosuki\gosuki.db`, and `127.0.0.1:2025`.
- Fork/debug GoSuki uses `D:\C2D\dotfiles\gosuki-dev`, including its own database, imports, logs, and `127.0.0.1:22025`.
- Debug launch scripts must validate the `gosuki-fork` role marker and must never stop or replace the Scoop executable.

## Windows debug workflow

- **Default Launch/Restart**: Always use `./restart-debug.ps1` to rebuild (`build-debug.ps1`) and reboot the debug executable on `127.0.0.1:22025`.
- **Stop**: `./stop-debug.ps1` to terminate only the debug process.
- **CLI**: `./run-gosuki.ps1` and `./run-suki.ps1` (defaults to fork runtime).
- Debug binary is `build\gosuki-debug.exe` (git-ignored).
