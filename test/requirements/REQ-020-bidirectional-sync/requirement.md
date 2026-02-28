# REQ-020: Git ↔ CVS Bidirectional Sync

## Description
Bidirectional synchronisation between a Git repository and a CVS repository
so that changes made in either system are reflected in the other.

## Acceptance Criteria
- [ ] New Git commits can be applied to a CVS repository (`git-to-cvs`)
- [ ] New CVS commits can be applied to a Git repository (`cvs-to-git`)
- [ ] Both directions can run in a single operation (`bidirectional`)
- [ ] Sync state is persisted so repeated runs only transfer new changes
- [ ] Dry-run mode previews planned changes without applying them
- [ ] CLI `sync` subcommand is available with `--config`, `--direction`, and `--dry-run` flags
- [ ] Sync configuration is loaded from a YAML file

## Test Cases
1. SyncDirection constants have correct string values
2. NewSyncer initialises all required fields
3. Sync state is saved to and loaded from a JSON file
4. Dry-run mode does not write the state file
5. Unknown sync direction returns an error
6. Git-to-CVS sync fails gracefully for a non-existent repository
7. CVS-to-Git sync fails gracefully for a non-existent repository
8. `loadSyncConfigFile` validates all required fields
9. Default direction is `bidirectional` when not specified in config
10. `printSyncInfo` does not panic with any valid config

## Status
✅ Complete
