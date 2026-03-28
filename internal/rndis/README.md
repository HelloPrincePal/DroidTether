# internal/rndis

Implementation of the Remote Network Driver Interface Specification (RNDIS) protocol over USB.

## Responsibility

- **State Machine**: Manages the lifecycle of an RNDIS session (INIT → QUERY → SET → DATA).
- **Messages**: Defines binary structures for all RNDIS message types.
- **Encoding**: Handles marshaling and unmarshaling of binary frames between host and device.
- **OIDs**: Manages Object Identifiers (OIDs) for querying MAC addresses, setting packet filters, etc.

## State Machine

1. **INITIALIZE**: Host sends `INITIALIZE_MSG`, device responds with `INITIALIZE_CMPLT`.
2. **QUERY (MAC)**: Host queries `OID_802_3_CURRENT_ADDRESS` to get the phone's MAC.
3. **SET (Filter)**: Host sets `OID_GEN_CURRENT_PACKET_FILTER` to enable data flow.
4. **DATA**: Device is now in data mode; Ethernet frames can flow on bulk endpoints.

## RNDIS over USB

Control messages are sent via USB Control Endpoint (EP0) using `SEND_ENCAPSULATED_COMMAND` and `GET_ENCAPSULATED_RESPONSE`. Data packets use Bulk IN/OUT endpoints.
