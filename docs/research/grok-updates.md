### Summary of CompuLink AutoStart System Learnings

From our conversation and research into NHRA/CompuLink rules (via web searches on official sources like NHRA.com and CompuLink docs), we've built a clear picture of the AutoStart system. It's an automated staging overseer in drag racing timing setups (e.g., CompuLink StarTrak), designed to prevent prolonged staging battles and ensure fair starts. Key mechanics include:

- **Core Workflow**:
    - **Arming**: Starter manually arms the system/tree (via switch), enabling monitoring. Beams can trigger pre-stage/stage lights even before arming for positioning, but AutoStart rules kick in only after.
    - **Activation/Three-Light Rule**: System activates when three top bulbs lit (both pre-stage + one stage across lanes). This starts monitoring; total bulbs >=3 (including deep staging where pre off but stage on counts).
    - **Timeout**: Once activated and first vehicle fully stages, the second has ~10s (default; 7-15s class-dependent) to stage—or red light foul for the second lane.
    - **Stability & Delay**: Both staged → 0.6s stability check (bulbs steady). Then random delay (0.6-1.1/1.4s + up to 0.2s variation, pro vs sportsman) before countdown sequence (ambers to green).
    - **Countdown Trigger**: Only on both staged (pre can be off for deep); tree executes pro (all ambers simultaneous) or sportsman (staggered).
    - **Faults/Overrides**: Timeout/guard beam fouls red light. Starter can override for bye runs. Courtesy staging encouraged (wait for both pre before staging), but not enforced—log violations.
    - **Class Variations**: Pro (shorter timeouts/delays, 0.4/0.5s green delay); Sportsman (longer, 10s default timeout).

Research confirmed no major 2025 changes—focus on fairness, with deep staging at driver's risk (no special wait).

### Differences From When We Started
Your initial tree model was solid for lights/sequence but blended staging and countdown, with duplicates (e.g., StartSequence vs StartStagingProcess) and no dedicated AutoStart logic. Key evolutions:
- **Separation of Concerns**: Started with tree handling everything (arming, activation, sequence). Now: Tree for lights/execution; AutoStart package for rules (activation, timeout, stability, events). Added integration.go for wiring (beams → autostart → tree/timing).
- **Rule Accuracy**: Initial three-light was strict (pre==2 + stage>=1); refined to total bulbs >=3 for deep/mixed cases. Timeout was global; now per-second-vehicle, lane-specific foul.
- **Events & Config**: No events initially; added bus for activation, timeout foul, tree trigger, fault, reset. Presets map for classes (Sportsman 10s timeout; ProFourTenths/ProFiveTenths with green delays noted for tree coord).
- **Tests/Edges**: Added deep staging tolerance, courtesy logs, class-specific configs. Tests verify timeouts cancel on timely stage, fault correct lane.
- **Defaults/Flexibility**: Shifted to sportsman default (10s timeout), with map-based presets for easy extension.

Overall, the implementation now closely matches real CompuLink/NHRA behavior, with better modularity and test coverage. If we missed anything, we can refine further!