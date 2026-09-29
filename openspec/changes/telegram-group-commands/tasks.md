# Tasks: Telegram group commands and friendlier onboarding

## 1. Group data layer
- [x] 1.1 Create `telegram/group.go` with `resolveCampaignContext` (bound enabled chat first, linked-user campaign fallback, binding/linking instructions otherwise)
- [x] 1.2 Implement party item query and `/items` renderer with quantity, notes, name ordering, cap and remainder footer
- [x] 1.3 Implement open quest query and `/quests` renderer (party characters, status `available`/`active`, grouped, objectives/rewards, caps, empty states)
- [x] 1.4 Implement location visit query and `/visits` renderer (newest first, character/location/type/relationship, cap, empty state)
- [x] 1.5 Implement campaign statistics query and renderer mirroring web analytics (characters, sessions, quests open/completed, member-owned NPCs and locations, average level)

## 2. Commands and context-aware statistics
- [x] 2.1 Register `/items`, `/quests`, `/visits` in the command registry and update their help entries
- [x] 2.2 Make `/stats` context-aware: campaign statistics in a bound campaign chat, character statistics otherwise, and note both modes in the help text

## 3. Onboarding
- [x] 3.1 Friendlier `/start` welcome (what the bot does, linking steps, post-link preview) with the navigation keyboard
- [x] 3.2 Rewrite the unlinked-user instructions with complete, friendly steps
- [x] 3.3 Post-link success message suggesting `/claim`, `/create`, and `/help`
- [x] 3.4 Subscribe to `my_chat_member` updates (client allowed updates), handle them in `bot.go`, and send the group-added welcome (connected campaign named when bound)

## 4. Tests
- [x] 4.1 Unit tests for campaign context resolution: bound hit, disabled row ignored, private fallback, unresolvable instructions
- [x] 4.2 Unit tests for `/items`, `/quests`, `/visits`, and campaign statistics: seeded data, caps, empty states, escaping, average level rounding
- [x] 4.3 Unit tests for onboarding: welcome, linking instructions, post-link guidance, group-added vs ordinary member update
- [x] 4.4 E2E: seed a party item, quest, and location visit via the HTTP API, bind the campaign chat, post `/items`, `/quests`, `/visits`, `/stats` webhook updates, and assert the replies
- [x] 4.5 Run `go test ./telegram/...`, `task lint:e2e`, and `npx tsc --noEmit` until green

## 5. Ship
- [ ] 5.1 Run the full local gates (`task ci`) until green
- [ ] 5.2 Commit with Conventional Commits, push the branch, open the PR with `gh pr create --fill`
- [ ] 5.3 Wait for checks (`gh pr checks --watch`); on red read `gh run view --log-failed`, fix, push, re-wait
- [ ] 5.4 Merge when green (`gh pr merge --squash --delete-branch`) and archive the OpenSpec change
