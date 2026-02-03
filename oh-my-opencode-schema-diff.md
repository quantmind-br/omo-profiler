# Diff do Schema oh-my-opencode

**Data:** 2026-02-03T15:12:04-03:00

## Comparação

```diff
--- oh-my-opencode-schema.json	2026-02-01 19:52:33.982674152 -0300
+++ /tmp/tmp.KjW2OYpSRP	2026-02-03 15:12:04.400276352 -0300
@@ -11,6 +11,9 @@
     "new_task_system_enabled": {
       "type": "boolean"
     },
+    "default_run_agent": {
+      "type": "string"
+    },
     "disabled_mcps": {
       "type": "array",
       "items": {
@@ -65,6 +68,7 @@
           "empty-task-response-detector",
           "think-mode",
           "anthropic-context-window-limit-recovery",
+          "preemptive-compaction",
           "rules-injector",
           "background-notification",
           "auto-update-checker",
@@ -2655,6 +2659,9 @@
         "auto_resume": {
           "type": "boolean"
         },
+        "preemptive_compaction": {
+          "type": "boolean"
+        },
         "truncate_all_tool_outputs": {
           "type": "boolean"
         },
@@ -2747,6 +2754,9 @@
               }
             }
           }
+        },
+        "task_system": {
+          "type": "boolean"
         }
       }
     },
```
