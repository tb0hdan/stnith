# stnith (STEE-NITH)

**DANGER: This is a highly destructive tool. Use with extreme caution.**

stnith is an advanced infrastructure security tool designed for the swift and irreversible destruction of data and hardware in compromised systems.
When triggered, it executes a pre-defined sequence of actions to erase data at rest, brick the hardware, and potentially initiate a physical destruction of components.

This tool is intended for use in high-security environments where the risk of data exfiltration from seized or compromised hardware is a critical concern.

## Key Features

- [ ] **Secure Data Erasure:** Overwrites and deletes data on storage devices to prevent recovery.
- [ ] **Failsafe Mechanisms:** Designed to be robust and challenging to neutralize once activated.
- [ ] **Hardware Bricking:** Corrupts firmware and critical system files to render the hardware unusable.
- [ ] **Physical Destruction:** Utilizes GPIO (General Purpose Input/Output) pins to send signals to external hardware, such as relays or custom circuits, to physically damage or destroy components.

## Disclaimer

This tool is intended for authorized use only. The authors are not responsible for any damage or loss of data resulting from the use or misuse of this software.
Ensure you have a complete understanding of its functionality and have implemented appropriate safeguards before deploying it in any environment.

## Usage

*WARNING: Due to safety reasons, binary versions of this software are not and will not be available. You must compile it yourself.*

Currently available with Dead Man's Switch functionality. Requires `root` to proceed.

1m before timer elapses:

```bash
stnith -dms 1m -disks
```

in the other console

```bash
stnith -reset
```

Once you've become used to it, you can enable destruction mode *at your own risk*:

```bash
stnith -dms 1w -disks -enable-it    
```
