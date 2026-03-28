# internal/daemon

The coordination layer of DroidTether.

## Responsibility

- **Daemon**: The main entry point that initializes the USB watcher and manages global state.
- **Session**: Orchestrates the lifecycle of a single connected phone. Each session owns its own `rndis`, `utun`, `dhcp`, and `relay` instances.
- **Lifecycle**: Handles transitions between `IDLE`, `CLAIMING`, `HANDSHAKING`, `ACTIVE`, and `TEARING DOWN`.

## Concurrency Model

- Main goroutine handles signals and global state.
- One goroutine for the USB watcher poll loop.
- Two goroutines per active session for bidirectional packet relaying.
