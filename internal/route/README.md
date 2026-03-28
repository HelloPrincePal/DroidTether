# internal/route

Manages macOS routing table entries to direct traffic through the `utun` interface.

## Responsibility

- **Default Route**: Injects a default route (`0.0.0.0/0`) pointing to the `utun` interface when a session becomes active.
- **Teardown**: Gracefully removes the default route when the phone is detached or the daemon stops.
- **Metric Management**: (Planned) Handles route priorities if multiple interfaces are present.

## Implementation

Currently uses system calls or `route` CLI commands to modify the macOS routing table.
