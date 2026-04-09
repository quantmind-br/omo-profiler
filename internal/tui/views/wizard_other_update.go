package views

import (
	"encoding/json"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/diogenes/omo-profiler/internal/config"
)

func (w WizardOther) Update(msg tea.Msg) (WizardOther, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		w.SetSize(msg.Width, msg.Height)
		return w, nil

	case tea.KeyMsg:
		if w.inSubSection && !w.subValueFocused {
			switch msg.String() {
			case "esc", "tab":
				w.inSubSection = false
				w.subValueFocused = false
				w.viewport.SetContent(w.renderContent())
				return w, nil
			case "up", "k":
				if w.subCursor > 0 {
					w.subCursor--
				}
				w.viewport.SetContent(w.renderContent())
				return w, nil
			case "down", "j":
				w.subCursor++
				w.viewport.SetContent(w.renderContent())
				return w, nil
			case " ":
				w.toggleSubItem()
				w.viewport.SetContent(w.renderContent())
				return w, nil
			case "enter", "right", "l":
				path := w.subSectionFieldPath(w.currentSection, w.subCursor)
				if path == "" && (w.currentSection != sectionOpenclaw || w.subCursor != 4) && (w.currentSection != sectionRuntimeFallback || w.subCursor != 1) && (w.currentSection != sectionSkillsJson || w.subCursor != 1) {
					return w, nil
				}
				w.subValueFocused = true
				w.viewport.SetContent(w.renderContent())
				return w, nil
			}
		}

		if w.inSubSection {
			if w.currentSection == sectionExperimental && w.subCursor == 6 {
				switch msg.String() {
				case "esc":
					w.dcpTurnProtTurns.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.dcpTurnProtTurns.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.dcpTurnProtTurns.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.dcpTurnProtTurns.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.dcpTurnProtTurns.Focus()
					w.dcpTurnProtTurns, cmd = w.dcpTurnProtTurns.Update(msg)
					return w, cmd
				default:
					w.dcpTurnProtTurns.Focus()
					w.dcpTurnProtTurns, cmd = w.dcpTurnProtTurns.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionExperimental && w.subCursor == 7 {
				switch msg.String() {
				case "esc":
					w.dcpProtectedTools.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.dcpProtectedTools.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.dcpProtectedTools.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.dcpProtectedTools.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.dcpProtectedTools.Focus()
					w.dcpProtectedTools, cmd = w.dcpProtectedTools.Update(msg)
					return w, cmd
				default:
					w.dcpProtectedTools.Focus()
					w.dcpProtectedTools, cmd = w.dcpProtectedTools.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionExperimental && w.subCursor == 12 {
				switch msg.String() {
				case "esc":
					w.dcpPurgeErrorsTurns.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.dcpPurgeErrorsTurns.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.dcpPurgeErrorsTurns.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.dcpPurgeErrorsTurns.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.dcpPurgeErrorsTurns.Focus()
					w.dcpPurgeErrorsTurns, cmd = w.dcpPurgeErrorsTurns.Update(msg)
					return w, cmd
				default:
					w.dcpPurgeErrorsTurns.Focus()
					w.dcpPurgeErrorsTurns, cmd = w.dcpPurgeErrorsTurns.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionExperimental && w.subCursor == 4 {
				switch msg.String() {
				case "right", "l":
					w.dcpNotificationIdx = (w.dcpNotificationIdx + 1) % len(dcpNotificationValues)
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "left", "h":
					w.dcpNotificationIdx = (w.dcpNotificationIdx - 1 + len(dcpNotificationValues)) % len(dcpNotificationValues)
					w.viewport.SetContent(w.renderContent())
					return w, nil
				}
			}

			if w.currentSection == sectionExperimental && w.subCursor == 15 {
				switch msg.String() {
				case "esc":
					w.expPluginLoadTimeoutMs.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.expPluginLoadTimeoutMs.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.expPluginLoadTimeoutMs.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.expPluginLoadTimeoutMs.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.expPluginLoadTimeoutMs.Focus()
					w.expPluginLoadTimeoutMs, cmd = w.expPluginLoadTimeoutMs.Update(msg)
					return w, cmd
				default:
					w.expPluginLoadTimeoutMs.Focus()
					w.expPluginLoadTimeoutMs, cmd = w.expPluginLoadTimeoutMs.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionExperimental && w.subCursor == 20 {
				switch msg.String() {
				case "esc":
					w.expMaxTools.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.expMaxTools.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.expMaxTools.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.expMaxTools.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.expMaxTools.Focus()
					w.expMaxTools, cmd = w.expMaxTools.Update(msg)
					return w, cmd
				default:
					w.expMaxTools.Focus()
					w.expMaxTools, cmd = w.expMaxTools.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionClaudeCode && w.subCursor == 6 {
				switch msg.String() {
				case "esc":
					w.ccPluginsOverride.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.ccPluginsOverride.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.ccPluginsOverride.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.ccPluginsOverride.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.ccPluginsOverride.Focus()
					w.ccPluginsOverride, cmd = w.ccPluginsOverride.Update(msg)
					return w, cmd
				default:
					w.ccPluginsOverride.Focus()
					w.ccPluginsOverride, cmd = w.ccPluginsOverride.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionBackgroundTask && w.subCursor == 0 {
				switch msg.String() {
				case "esc":
					w.btProviderConcurrency.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.btProviderConcurrency.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.btProviderConcurrency.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.btProviderConcurrency.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.btProviderConcurrency.Focus()
					w.btProviderConcurrency, cmd = w.btProviderConcurrency.Update(msg)
					return w, cmd
				default:
					w.btProviderConcurrency.Focus()
					w.btProviderConcurrency, cmd = w.btProviderConcurrency.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionBackgroundTask && w.subCursor == 1 {
				switch msg.String() {
				case "esc":
					w.btModelConcurrency.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.btModelConcurrency.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.btModelConcurrency.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.btModelConcurrency.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.btModelConcurrency.Focus()
					w.btModelConcurrency, cmd = w.btModelConcurrency.Update(msg)
					return w, cmd
				default:
					w.btModelConcurrency.Focus()
					w.btModelConcurrency, cmd = w.btModelConcurrency.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionDisabledMcps {
				switch msg.String() {
				case "esc":
					w.disabledMcps.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.disabledMcps.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				default:
					w.disabledMcps.Focus()
					w.disabledMcps, cmd = w.disabledMcps.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionDisabledTools {
				switch msg.String() {
				case "esc":
					w.disabledTools.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.disabledTools.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				default:
					w.disabledTools.Focus()
					w.disabledTools, cmd = w.disabledTools.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionRalphLoop && w.subCursor == 1 {
				switch msg.String() {
				case "esc":
					w.rlDefaultMaxIterations.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.rlDefaultMaxIterations.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.rlDefaultMaxIterations.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.rlDefaultMaxIterations.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.rlDefaultMaxIterations.Focus()
					w.rlDefaultMaxIterations, cmd = w.rlDefaultMaxIterations.Update(msg)
					return w, cmd
				default:
					w.rlDefaultMaxIterations.Focus()
					w.rlDefaultMaxIterations, cmd = w.rlDefaultMaxIterations.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionRalphLoop && w.subCursor == 2 {
				switch msg.String() {
				case "esc":
					w.rlStateDir.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.rlStateDir.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.rlStateDir.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.rlStateDir.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.rlStateDir.Focus()
					w.rlStateDir, cmd = w.rlStateDir.Update(msg)
					return w, cmd
				default:
					w.rlStateDir.Focus()
					w.rlStateDir, cmd = w.rlStateDir.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionRalphLoop && w.subCursor == 3 {
				switch msg.String() {
				case "right", "l":
					w.rlDefaultStrategyIdx = (w.rlDefaultStrategyIdx + 1) % len(ralphLoopStrategies)
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "left", "h":
					w.rlDefaultStrategyIdx = (w.rlDefaultStrategyIdx - 1 + len(ralphLoopStrategies)) % len(ralphLoopStrategies)
					w.viewport.SetContent(w.renderContent())
					return w, nil
				}
			}

			if w.currentSection == sectionBackgroundTask && w.subCursor == 2 {
				switch msg.String() {
				case "esc":
					w.btDefaultConcurrency.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.btDefaultConcurrency.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.btDefaultConcurrency.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.btDefaultConcurrency.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.btDefaultConcurrency.Focus()
					w.btDefaultConcurrency, cmd = w.btDefaultConcurrency.Update(msg)
					return w, cmd
				default:
					w.btDefaultConcurrency.Focus()
					w.btDefaultConcurrency, cmd = w.btDefaultConcurrency.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionBackgroundTask && w.subCursor == 3 {
				switch msg.String() {
				case "esc":
					w.btStaleTimeoutMs.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.btStaleTimeoutMs.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.btStaleTimeoutMs.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.btStaleTimeoutMs.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.btStaleTimeoutMs.Focus()
					w.btStaleTimeoutMs, cmd = w.btStaleTimeoutMs.Update(msg)
					return w, cmd
				default:
					w.btStaleTimeoutMs.Focus()
					w.btStaleTimeoutMs, cmd = w.btStaleTimeoutMs.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionBackgroundTask && w.subCursor == 4 {
				switch msg.String() {
				case "esc":
					w.btMessageStalenessTimeoutMs.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.btMessageStalenessTimeoutMs.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.btMessageStalenessTimeoutMs.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.btMessageStalenessTimeoutMs.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.btMessageStalenessTimeoutMs.Focus()
					w.btMessageStalenessTimeoutMs, cmd = w.btMessageStalenessTimeoutMs.Update(msg)
					return w, cmd
				default:
					w.btMessageStalenessTimeoutMs.Focus()
					w.btMessageStalenessTimeoutMs, cmd = w.btMessageStalenessTimeoutMs.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionBackgroundTask && w.subCursor == 5 {
				switch msg.String() {
				case "esc":
					w.btSyncPollTimeoutMs.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.btSyncPollTimeoutMs.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.btSyncPollTimeoutMs.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.btSyncPollTimeoutMs.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.btSyncPollTimeoutMs.Focus()
					w.btSyncPollTimeoutMs, cmd = w.btSyncPollTimeoutMs.Update(msg)
					return w, cmd
				default:
					w.btSyncPollTimeoutMs.Focus()
					w.btSyncPollTimeoutMs, cmd = w.btSyncPollTimeoutMs.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionBackgroundTask && w.subCursor == 6 {
				switch msg.String() {
				case "esc":
					w.btMaxDepth.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.btMaxDepth.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.btMaxDepth.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.btMaxDepth.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.btMaxDepth.Focus()
					w.btMaxDepth, cmd = w.btMaxDepth.Update(msg)
					return w, cmd
				default:
					w.btMaxDepth.Focus()
					w.btMaxDepth, cmd = w.btMaxDepth.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionBackgroundTask && w.subCursor == 7 {
				switch msg.String() {
				case "esc":
					w.btMaxDescendants.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.btMaxDescendants.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.btMaxDescendants.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.btMaxDescendants.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.btMaxDescendants.Focus()
					w.btMaxDescendants, cmd = w.btMaxDescendants.Update(msg)
					return w, cmd
				default:
					w.btMaxDescendants.Focus()
					w.btMaxDescendants, cmd = w.btMaxDescendants.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionBackgroundTask && w.subCursor == 8 {
				switch msg.String() {
				case "esc":
					w.btTaskTtlMs.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.btTaskTtlMs.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.btTaskTtlMs.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.btTaskTtlMs.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.btTaskTtlMs.Focus()
					w.btTaskTtlMs, cmd = w.btTaskTtlMs.Update(msg)
					return w, cmd
				default:
					w.btTaskTtlMs.Focus()
					w.btTaskTtlMs, cmd = w.btTaskTtlMs.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionBackgroundTask && w.subCursor == 9 {
				switch msg.String() {
				case "esc":
					w.btSessionGoneTimeoutMs.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.btSessionGoneTimeoutMs.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.btSessionGoneTimeoutMs.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.btSessionGoneTimeoutMs.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.btSessionGoneTimeoutMs.Focus()
					w.btSessionGoneTimeoutMs, cmd = w.btSessionGoneTimeoutMs.Update(msg)
					return w, cmd
				default:
					w.btSessionGoneTimeoutMs.Focus()
					w.btSessionGoneTimeoutMs, cmd = w.btSessionGoneTimeoutMs.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionBackgroundTask && w.subCursor == 10 {
				switch msg.String() {
				case "esc":
					w.btMaxToolCalls.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.btMaxToolCalls.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.btMaxToolCalls.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.btMaxToolCalls.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.btMaxToolCalls.Focus()
					w.btMaxToolCalls, cmd = w.btMaxToolCalls.Update(msg)
					return w, cmd
				default:
					w.btMaxToolCalls.Focus()
					w.btMaxToolCalls, cmd = w.btMaxToolCalls.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionBackgroundTask && w.subCursor == 12 {
				switch msg.String() {
				case "esc":
					w.btCircuitBreakerMaxCalls.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.btCircuitBreakerMaxCalls.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.btCircuitBreakerMaxCalls.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.btCircuitBreakerMaxCalls.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.btCircuitBreakerMaxCalls.Focus()
					w.btCircuitBreakerMaxCalls, cmd = w.btCircuitBreakerMaxCalls.Update(msg)
					return w, cmd
				default:
					w.btCircuitBreakerMaxCalls.Focus()
					w.btCircuitBreakerMaxCalls, cmd = w.btCircuitBreakerMaxCalls.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionBackgroundTask && w.subCursor == 13 {
				switch msg.String() {
				case "esc":
					w.btCircuitBreakerConsecutive.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.btCircuitBreakerConsecutive.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.btCircuitBreakerConsecutive.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.btCircuitBreakerConsecutive.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.btCircuitBreakerConsecutive.Focus()
					w.btCircuitBreakerConsecutive, cmd = w.btCircuitBreakerConsecutive.Update(msg)
					return w, cmd
				default:
					w.btCircuitBreakerConsecutive.Focus()
					w.btCircuitBreakerConsecutive, cmd = w.btCircuitBreakerConsecutive.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionGitMaster && w.subCursor == 1 {
				switch msg.String() {
				case "esc":
					w.gmCommitFooterText.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.gmCommitFooterText.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.gmCommitFooterText.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.gmCommitFooterText.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.gmCommitFooterText.Focus()
					w.gmCommitFooterText, cmd = w.gmCommitFooterText.Update(msg)
					return w, cmd
				default:
					w.gmCommitFooterText.Focus()
					w.gmCommitFooterText, cmd = w.gmCommitFooterText.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionGitMaster && w.subCursor == 3 {
				switch msg.String() {
				case "esc":
					w.gmGitEnvPrefix.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.gmGitEnvPrefix.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.gmGitEnvPrefix.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.gmGitEnvPrefix.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.gmGitEnvPrefix.Focus()
					w.gmGitEnvPrefix, cmd = w.gmGitEnvPrefix.Update(msg)
					return w, cmd
				default:
					w.gmGitEnvPrefix.Focus()
					w.gmGitEnvPrefix, cmd = w.gmGitEnvPrefix.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionTmux && w.subCursor == 5 {
				switch msg.String() {
				case "right", "l":
					w.tmuxIsolationIdx = (w.tmuxIsolationIdx + 1) % len(tmuxIsolations)
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "left", "h":
					w.tmuxIsolationIdx = (w.tmuxIsolationIdx - 1 + len(tmuxIsolations)) % len(tmuxIsolations)
					w.viewport.SetContent(w.renderContent())
					return w, nil
				}
			}

			if w.currentSection == sectionStartWork {
				switch msg.String() {
				case "esc", "tab":
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.startWorkAutoCommit = !w.startWorkAutoCommit
					w.viewport.SetContent(w.renderContent())
					return w, nil
				}
			}

			if w.currentSection == sectionCommentChecker {
				switch msg.String() {
				case "esc":
					w.ccCustomPrompt.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.ccCustomPrompt.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				default:
					w.ccCustomPrompt.Focus()
					w.ccCustomPrompt, cmd = w.ccCustomPrompt.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionBabysitting {
				switch msg.String() {
				case "esc":
					w.babysittingTimeoutMs.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.babysittingTimeoutMs.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.babysittingTimeoutMs.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.babysittingTimeoutMs.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.babysittingTimeoutMs.Focus()
					w.babysittingTimeoutMs, cmd = w.babysittingTimeoutMs.Update(msg)
					return w, cmd
				default:
					w.babysittingTimeoutMs.Focus()
					w.babysittingTimeoutMs, cmd = w.babysittingTimeoutMs.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionBrowserAutomationEngine && w.subCursor == 0 {
				switch msg.String() {
				case "right", "l":
					w.browserProviderIdx = (w.browserProviderIdx + 1) % len(browserProviders)
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "left", "h":
					w.browserProviderIdx = (w.browserProviderIdx - 1 + len(browserProviders)) % len(browserProviders)
					w.viewport.SetContent(w.renderContent())
					return w, nil
				}
			}

			if w.currentSection == sectionTmux && w.subCursor == 1 {
				switch msg.String() {
				case "right", "l":
					w.tmuxLayoutIdx = (w.tmuxLayoutIdx + 1) % len(tmuxLayouts)
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "left", "h":
					w.tmuxLayoutIdx = (w.tmuxLayoutIdx - 1 + len(tmuxLayouts)) % len(tmuxLayouts)
					w.viewport.SetContent(w.renderContent())
					return w, nil
				}
			}

			if w.currentSection == sectionTmux && w.subCursor == 2 {
				switch msg.String() {
				case "esc":
					w.tmuxMainPaneSize.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.tmuxMainPaneSize.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.tmuxMainPaneSize.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.tmuxMainPaneSize.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.tmuxMainPaneSize.Focus()
					w.tmuxMainPaneSize, cmd = w.tmuxMainPaneSize.Update(msg)
					return w, cmd
				default:
					w.tmuxMainPaneSize.Focus()
					w.tmuxMainPaneSize, cmd = w.tmuxMainPaneSize.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionTmux && w.subCursor == 3 {
				switch msg.String() {
				case "esc":
					w.tmuxMainPaneMinWidth.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.tmuxMainPaneMinWidth.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.tmuxMainPaneMinWidth.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.tmuxMainPaneMinWidth.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.tmuxMainPaneMinWidth.Focus()
					w.tmuxMainPaneMinWidth, cmd = w.tmuxMainPaneMinWidth.Update(msg)
					return w, cmd
				default:
					w.tmuxMainPaneMinWidth.Focus()
					w.tmuxMainPaneMinWidth, cmd = w.tmuxMainPaneMinWidth.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionTmux && w.subCursor == 4 {
				switch msg.String() {
				case "esc":
					w.tmuxAgentPaneMinWidth.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.tmuxAgentPaneMinWidth.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.tmuxAgentPaneMinWidth.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.tmuxAgentPaneMinWidth.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.tmuxAgentPaneMinWidth.Focus()
					w.tmuxAgentPaneMinWidth, cmd = w.tmuxAgentPaneMinWidth.Update(msg)
					return w, cmd
				default:
					w.tmuxAgentPaneMinWidth.Focus()
					w.tmuxAgentPaneMinWidth, cmd = w.tmuxAgentPaneMinWidth.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionWebsearch && w.subCursor == 0 {
				switch msg.String() {
				case "right", "l":
					w.websearchProviderIdx = (w.websearchProviderIdx + 1) % len(websearchProviders)
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "left", "h":
					w.websearchProviderIdx = (w.websearchProviderIdx - 1 + len(websearchProviders)) % len(websearchProviders)
					w.viewport.SetContent(w.renderContent())
					return w, nil
				}
			}

			if w.currentSection == sectionSisyphus && w.subCursor == 0 {
				switch msg.String() {
				case "esc":
					w.sisyphusTasksStoragePath.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.sisyphusTasksStoragePath.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.sisyphusTasksStoragePath.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.sisyphusTasksStoragePath.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.sisyphusTasksStoragePath.Focus()
					w.sisyphusTasksStoragePath, cmd = w.sisyphusTasksStoragePath.Update(msg)
					return w, cmd
				default:
					w.sisyphusTasksStoragePath.Focus()
					w.sisyphusTasksStoragePath, cmd = w.sisyphusTasksStoragePath.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionSisyphus && w.subCursor == 1 {
				switch msg.String() {
				case "esc":
					w.sisyphusTasksTaskListID.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.sisyphusTasksTaskListID.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.sisyphusTasksTaskListID.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.sisyphusTasksTaskListID.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.sisyphusTasksTaskListID.Focus()
					w.sisyphusTasksTaskListID, cmd = w.sisyphusTasksTaskListID.Update(msg)
					return w, cmd
				default:
					w.sisyphusTasksTaskListID.Focus()
					w.sisyphusTasksTaskListID, cmd = w.sisyphusTasksTaskListID.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionDefaultRunAgent {
				switch msg.String() {
				case "esc":
					w.defaultRunAgent.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.defaultRunAgent.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				default:
					w.defaultRunAgent.Focus()
					w.defaultRunAgent, cmd = w.defaultRunAgent.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionModelCapabilities && w.subCursor == 2 {
				switch msg.String() {
				case "esc":
					w.mcRefreshTimeoutMs.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.mcRefreshTimeoutMs.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.mcRefreshTimeoutMs.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.mcRefreshTimeoutMs.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.mcRefreshTimeoutMs.Focus()
					w.mcRefreshTimeoutMs, cmd = w.mcRefreshTimeoutMs.Update(msg)
					return w, cmd
				default:
					w.mcRefreshTimeoutMs.Focus()
					w.mcRefreshTimeoutMs, cmd = w.mcRefreshTimeoutMs.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionModelCapabilities && w.subCursor == 3 {
				switch msg.String() {
				case "esc":
					w.mcSourceURL.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.mcSourceURL.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.mcSourceURL.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.mcSourceURL.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.mcSourceURL.Focus()
					w.mcSourceURL, cmd = w.mcSourceURL.Update(msg)
					return w, cmd
				default:
					w.mcSourceURL.Focus()
					w.mcSourceURL, cmd = w.mcSourceURL.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionOpenclaw && w.subCursor == 4 {
				switch msg.String() {
				case "esc":
					w.openclawEditor.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.openclawEditor.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				default:
					w.openclawEditor.Focus()
					w.openclawEditor, cmd = w.openclawEditor.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionRuntimeFallback && w.subCursor == 1 {
				switch msg.String() {
				case "esc":
					w.runtimeFallbackEditor.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.runtimeFallbackEditor.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				default:
					w.runtimeFallbackEditor.Focus()
					w.runtimeFallbackEditor, cmd = w.runtimeFallbackEditor.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionSkillsJson && w.subCursor == 1 {
				switch msg.String() {
				case "esc":
					w.skillsEditor.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.skillsEditor.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				default:
					w.skillsEditor.Focus()
					w.skillsEditor, cmd = w.skillsEditor.Update(msg)
					return w, cmd
				}
			}

			switch msg.String() {

			case "esc":
				w.inSubSection = false
			case "up", "k":
				if w.subCursor > 0 {
					w.subCursor--
				}
			case "down", "j":
				w.subCursor++
			case " ", "enter":
				w.toggleSubItem()
			case "tab":
				w.inSubSection = false
			}
			w.viewport.SetContent(w.renderContent())
			return w, nil
		}

		switch {
		case key.Matches(msg, w.keys.Up):
			w.simpleValueFocused = false
			if w.currentSection > 0 {
				w.currentSection--
			}
		case key.Matches(msg, w.keys.Down):
			w.simpleValueFocused = false
			if w.currentSection < sectionSkillsJson {
				w.currentSection++
			}
		case key.Matches(msg, w.keys.Toggle):
			w.toggleSection()
		case key.Matches(msg, w.keys.Expand):
			w.sectionExpanded[w.currentSection] = !w.sectionExpanded[w.currentSection]
			if w.sectionExpanded[w.currentSection] {
				w.inSubSection = true
				w.subCursor = 0
				w.subValueFocused = false
			}
		case key.Matches(msg, w.keys.Right):
			if w.isSimpleBooleanSection(w.currentSection) {
				w.simpleValueFocused = true
				break
			}
			if !w.inSubSection && !w.sectionExpanded[w.currentSection] {
				w.sectionExpanded[w.currentSection] = true
				w.inSubSection = true
				w.subCursor = 0
				w.subValueFocused = false
			}
		case key.Matches(msg, w.keys.Left):
			if w.isSimpleBooleanSection(w.currentSection) {
				w.simpleValueFocused = false
				break
			}
			if !w.inSubSection && w.sectionExpanded[w.currentSection] {
				w.sectionExpanded[w.currentSection] = false
			}
		case key.Matches(msg, w.keys.Next):
			return w, func() tea.Msg { return WizardNextMsg{} }
		case key.Matches(msg, w.keys.Back):
			return w, func() tea.Msg { return WizardBackMsg{} }
		}
	}

	// Update viewport
	w.viewport.SetContent(w.renderContent())
	w.viewport, cmd = w.viewport.Update(msg)

	return w, cmd
}

func (w *WizardOther) toggleSection() {
	switch w.currentSection {
	case sectionAutoUpdate:
		if w.simpleValueFocused {
			w.autoUpdate = !w.autoUpdate
		} else {
			w.toggleFieldSelection(autoUpdateFieldPath)
		}
	case sectionNewTaskSystemEnabled:
		if w.simpleValueFocused {
			w.newTaskSystemEnabled = !w.newTaskSystemEnabled
		} else {
			w.toggleFieldSelection(newTaskSystemEnabledFieldPath)
		}
	case sectionHashlineEdit:
		if w.simpleValueFocused {
			w.hashlineEdit = !w.hashlineEdit
		} else {
			w.toggleFieldSelection(hashlineEditFieldPath)
		}
	case sectionModelFallback:
		if w.simpleValueFocused {
			w.modelFallback = !w.modelFallback
		} else {
			w.toggleFieldSelection(modelFallbackFieldPath)
		}
	case sectionDisabledMcps:
		w.toggleFieldSelection(w.topLevelFieldPath(sectionDisabledMcps))
	case sectionDisabledAgents:
		w.toggleFieldSelection(w.topLevelFieldPath(sectionDisabledAgents))
	case sectionDisabledSkills:
		w.toggleFieldSelection(w.topLevelFieldPath(sectionDisabledSkills))
	case sectionDisabledCommands:
		w.toggleFieldSelection(w.topLevelFieldPath(sectionDisabledCommands))
	case sectionDisabledTools:
		w.toggleFieldSelection(w.topLevelFieldPath(sectionDisabledTools))
	}
}

func (w *WizardOther) toggleSubItem() {
	// Handle Row 0 inclusion toggle for disabled-list sections
	if w.subCursor == 0 {
		switch w.currentSection {
		case sectionDisabledMcps, sectionDisabledAgents, sectionDisabledSkills,
			sectionDisabledCommands, sectionDisabledTools:
			w.toggleFieldSelection(w.topLevelFieldPath(w.currentSection))
			return
		}
	}

	path := w.subSectionFieldPath(w.currentSection, w.subCursor)
	if path != "" && !w.subValueFocused {
		w.toggleFieldSelection(path)
		return
	}

	switch w.currentSection {
	case sectionDisabledAgents:
		idx := w.subCursor - 1
		if idx >= 0 && idx < len(disableableAgents) {
			agent := disableableAgents[idx]
			w.disabledAgents[agent] = !w.disabledAgents[agent]
		}
	case sectionDisabledSkills:
		idx := w.subCursor - 1
		if idx >= 0 && idx < len(disableableSkills) {
			skill := disableableSkills[idx]
			w.disabledSkills[skill] = !w.disabledSkills[skill]
		}
	case sectionDisabledCommands:
		idx := w.subCursor - 1
		if idx >= 0 && idx < len(disableableCommands) {
			cmd := disableableCommands[idx]
			w.disabledCommands[cmd] = !w.disabledCommands[cmd]
		}
	case sectionExperimental:
		switch w.subCursor {
		case 0:
			w.expAggressiveTrunc = !w.expAggressiveTrunc
		case 1:
			w.expAutoResume = !w.expAutoResume
		case 2:
			w.expTruncateAllOutputs = !w.expTruncateAllOutputs
		case 3:
			w.dcpEnabled = !w.dcpEnabled
		case 5:
			w.dcpTurnProtEnabled = !w.dcpTurnProtEnabled
		case 8:
			w.dcpDeduplicationEnabled = !w.dcpDeduplicationEnabled
		case 9:
			w.dcpSupersedeWritesEnabled = !w.dcpSupersedeWritesEnabled
		case 10:
			w.dcpSupersedeWritesAggressive = !w.dcpSupersedeWritesAggressive
		case 11:
			w.dcpPurgeErrorsEnabled = !w.dcpPurgeErrorsEnabled
		case 13:
			w.expPreemptiveCompaction = !w.expPreemptiveCompaction
		case 14:
			w.expTaskSystem = !w.expTaskSystem
		case 16:
			w.expSafeHookCreation = !w.expSafeHookCreation
		case 17:
			w.expHashlineEdit = !w.expHashlineEdit
		case 18:
			w.expDisableOmoEnv = !w.expDisableOmoEnv
		case 19:
			w.expModelFallbackTitle = !w.expModelFallbackTitle
		}
	case sectionClaudeCode:
		switch w.subCursor {
		case 0:
			w.ccMcp = !w.ccMcp
		case 1:
			w.ccCommands = !w.ccCommands
		case 2:
			w.ccSkills = !w.ccSkills
		case 3:
			w.ccAgents = !w.ccAgents
		case 4:
			w.ccHooks = !w.ccHooks
		case 5:
			w.ccPlugins = !w.ccPlugins
		}
	case sectionSisyphusAgent:
		switch w.subCursor {
		case 0:
			w.saDisabled = !w.saDisabled
		case 1:
			w.saDefaultBuilderEnabled = !w.saDefaultBuilderEnabled
		case 2:
			w.saPlannerEnabled = !w.saPlannerEnabled
		case 3:
			w.saReplacePlan = !w.saReplacePlan
		case 4:
			w.saTDD = !w.saTDD
		}
	case sectionRalphLoop:
		if w.subCursor == 0 {
			w.rlEnabled = !w.rlEnabled
		}
	case sectionNotification:
		if w.subCursor == 0 {
			w.notifForceEnable = !w.notifForceEnable
		}
	case sectionGitMaster:
		switch w.subCursor {
		case 0:
			w.gmCommitFooter = !w.gmCommitFooter
		case 2:
			w.gmIncludeCoAuthoredBy = !w.gmIncludeCoAuthoredBy
		}
	case sectionTmux:
		if w.subCursor == 0 {
			w.tmuxEnabled = !w.tmuxEnabled
		}
	case sectionSisyphus:
		if w.subCursor == 2 {
			w.sisyphusTasksClaudeCodeCompat = !w.sisyphusTasksClaudeCodeCompat
		}
	case sectionStartWork:
		if w.subCursor == 0 {
			w.startWorkAutoCommit = !w.startWorkAutoCommit
		}
	case sectionModelCapabilities:
		switch w.subCursor {
		case 0:
			w.mcEnabled = !w.mcEnabled
		case 1:
			w.mcAutoRefreshOnStart = !w.mcAutoRefreshOnStart
		}
	case sectionOpenclaw:
		if w.subCursor == 0 {
			var parsed config.OpenclawConfig
			if err := json.Unmarshal([]byte(w.openclawEditor.Value()), &parsed); err == nil && parsed.Enabled != nil {
				parsed.Enabled = wizardOtherBoolPtr(!*parsed.Enabled)
				if raw, err := json.MarshalIndent(parsed, "", "  "); err == nil {
					w.openclawEditor.SetValue(string(raw))
				}
			}
		}
	}
}

