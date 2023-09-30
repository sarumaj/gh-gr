package resources

type SecurityAndAnalysis struct {
	AdvancedSecurity             AdvancedSecurity             `json:"advanced_security"`
	SecretScanning               SecretScanning               `json:"secret_scanning"`
	SecretScanningPushProtection SecretScanningPushProtection `json:"secret_scanning_push_protection"`
}
