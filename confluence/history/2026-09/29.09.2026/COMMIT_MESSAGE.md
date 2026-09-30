# 1) Drop the interrupted cherry-pick (keeps your working-tree edits)
git cherry-pick --abort

# 2) Stage the wave (games + DTO + goods + notes)
git add \
  frontend/src/App.tsx \
  frontend/src/components/Home/Home.tsx \
  frontend/src/components/Layout/AppNavbar.tsx \
  frontend/src/components/Leaderboard/Leaderboard.tsx \
  frontend/src/components/Leaderboard/LeaderboardFull.tsx \
  frontend/src/components/Profile/Profile.tsx \
  frontend/src/components/Inventory/InventoryCreateModal.tsx \
  frontend/src/components/Inventory/InventoryPage.tsx \
  frontend/src/components/Games/Twenty48 \
  frontend/src/components/Games/Gears \
  frontend/src/components/Games/Companion \
  services/game/go.mod \
  services/game/go.sum \
  services/game/internal/service/game_service.go \
  services/gateway/internal/app/gateway.go \
  services/gateway/internal/dto/twins.go \
  services/gateway/api/openapi.yaml \
  services/inventory/internal/model/item.go \
  docs/openapi.yaml \
  confluence/history/2026-09/29.09.2026/

# optional: if IDEAS.md delete is intentional
git add -u confluence/history/2026-09/29.09.2026/IDEAS.md

# skip for now unless you want it in this commit:
# confluence/history/2026-09/30.09.2026/

# 3) Commit
git commit -m "$(cat <<'EOF'
feat: 8-game pool, twin gin.H DTOs, KKI card goods, burger polish

Add twenty48/gears/companion, map list endpoints via dto, and inventory type карточка.
EOF
)"

# 4) Check
git status