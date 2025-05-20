# yaml-helm-pipeline

Allow users to trigger Helm-based secret generation through a UI, instead of giving direct file access, and then commit the resulting YAML to GitHub.

## TODO

rchitecture Components
	1.	Frontend UI
	•	SolidJS
	•	Form input (if values.yaml is user-customizable).
	•	Trigger button to initiate the secret generation.
	2.	Backend API Service (Golang)
	•	Receives the trigger from UI.
	•	Runs Helm CLI with a templated chart.
	•	Handles Git operations.
	•	Authenticates with GitHub (token-based).
	•	Validates inputs and applies audit logging.
	3.	GitHub Integration
	•	GitHub PAT (Personal Access Token) stored securely.
	•	Uses a bot account or service user.
	4.	Helm CLI or Library
	•	Run helm template programmatically using:
	•	Shell out to helm binary, or
	•	Use a Go Helm SDK.
    5. Allow user to preview changing keys without values from the ui before committing and pushing.