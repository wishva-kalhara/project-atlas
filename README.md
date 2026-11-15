1. `./tools.exe new-workflow-fbScheduling`
2. `./harness.exe -url "https://gemini.google.com/gem/9bfe84b773a7/a4f05b58db4c8827"`
3. `./tools.exe gen-workflow-googleflow-in`
4. `./workflow-exec.exe video-gen --headed --index=0`
5. `./tools.exe gen-workflow-fbSchedule-in --fbPageSlug "marine.masterpieces" --pageName "Marine Masterpieces"`
6. `./workflow-exec.exe schedule-fb --headed`
7. `./tools.exe clear-tmp-data`
