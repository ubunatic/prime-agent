.PHONY: help test check verify status interactive-test

help:
	@echo "Available make targets:"
	@echo "  make check             - Run formatting, linting, and static checks"
	@echo "  make verify            - Assert fork invariant integrity"
	@echo "  make status            - Check upstream sync status"
	@echo "  make test              - Run package unit tests"
	@echo "  make interactive-test  - Run isolated TUI smoke test in tmux (offline, no API tokens spent)"

check:
	npm run check

verify:
	npm run fork-sync -- verify --run-checks

status:
	npm run fork-sync -- status

test:
	npm test

interactive-test:
	@echo "[*] Launching Prime Agent TUI smoke test in isolated tmux session..."
	@SESSION="prime-agent-test-$$$$"; \
	tmux new-session -d -s "$$SESSION" -x 100 -y 30 "cd $(CURDIR) && ./prime-agent.sh --offline --no-session" && \
	sleep 2 && \
	tmux capture-pane -t "$$SESSION" -p && \
	tmux kill-session -t "$$SESSION" && \
	printf "\n[✓] Interactive TUI smoke test completed.\n"
