# fx-resources-watch

Stdio MCP fixture server exposing one subscribable resource and emitting
deterministic `notifications/resources/updated` events.

## Capabilities

| Capability | Value |
|---|---|
| `resources.subscribe` | `true` |

No tools. The fixture deliberately has no tools so the harness cannot confuse
tool round-trips with subscription event behaviour.

## Resources

| URI | MIME | Read body |
|---|---|---|
| `test://changes` | `text/plain` | `"fx-resources-watch test resource"` |

## Flags

| Flag | Default | Meaning |
|---|---|---|
| `--event-interval` | `200ms` | Cadence between emitted events. |
| `--event-count` | `5` | Number of events to emit. `-1` = emit forever. |

After the client subscribes to `test://changes`, the server starts a ticker
that emits `notifications/resources/updated` every `--event-interval`. After
`--event-count` events (when `>= 0`), the server cancels its main context and
the stdio transport closes cleanly, so the client observes EOF and exits 0.

## Event payload

The standard `ResourceUpdatedNotificationParams` carries only `uri` at the
top level. The fixture's per-event sequence number is therefore placed in
the protocol-reserved `_meta` field:

    {
      "jsonrpc": "2.0",
      "method": "notifications/resources/updated",
      "params": {
        "uri": "test://changes",
        "_meta": { "seq": 1 }
      }
    }

The harness asserts payload equality (URI + `_meta.seq`) and total event
count, NOT inter-event timing — the schedule is wall-clock-driven.

## Backs

- FX-SIGINT-02 — spawn with `--event-count=-1`; the harness SIGINTs the
  client after the first event and expects the client to unsubscribe and
  exit 130.
- FX-SIGINT-03 — spawn with `--event-count=5 --event-interval=100ms`; the
  client receives all 5 events, the server closes its end of the connection,
  the client exits 0.

## Notes

The ticker is started inside the SubscribeHandler under a `sync.Once`. A
client that subscribes a second time won't double-fire the schedule; the
single `--event-count` budget still applies.
