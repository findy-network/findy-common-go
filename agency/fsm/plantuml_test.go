package fsm

import (
	"net/url"
	"testing"
)

func TestGenerateURL(t *testing.T) {
	simpleURL, _ := url.Parse("http://www.plantuml.com/plantuml/svg/AiaioKbLo4rCpiZCI-MArefLqBLJy7JnSUKgBaaiILLGKY0HWFIIgaT98R4OOYbOjwwkdKAuesU8fvzxV728OqXei9M2bbPIOd5cSdnkQd5nOdfgjOAIKgsMLanUDQgmZIv8ebR1rjOk9e-B33-Wsakg7r1rSw4PfWiDMapVG1LWznD24kulG0000F__")
	proofMachineURL, _ := url.Parse("http://www.plantuml.com/plantuml/svg/AiaioKbLA2ZApqzJo4rCpiZCI-MArefLqBLJy7JnSUKgBaaiILLGKY0HWFIIgaT98R4OOYbOjwwkdKAuesU8fvzxV728OqXei9M2bbPoVbvUQd99PdvUjOAIKgsMLanUTL9YSMPoV6vgSN5YUcgrZIv8ebR1rjOk9e-B30-WMagg1r1rSw4PfWiDLv1NK9qDLO3TJmX9kBy00G00__y")
	type args struct {
		subPath string
		m       *Machine
	}
	tests := []struct {
		name    string
		args    args
		wantURL *url.URL
		wantErr bool
	}{
		{name: "simple", args: args{"svg", &machine}, wantErr: false,
			wantURL: simpleURL},
		{name: "simpleTerminate", args: args{"svg", &machineTerminates}, wantErr: false,
			wantURL: simpleURL},
		{name: "proof machine", args: args{"svg", &showProofMachine}, wantErr: false,
			wantURL: proofMachineURL},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotURL, err := GenerateURL(tt.args.subPath, tt.args.m)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// not very good test
			if gotURL == nil || gotURL.Host != tt.wantURL.Host || !gotURL.IsAbs() {
				t.Errorf("GenerateURL() gotURL = %v, want %v", gotURL, tt.wantURL)
			}
		})
	}
}
