package models

import "encoding/xml"

/*
	Generated from https://www.onlinetool.io/xmltogo/ based on what I got from
	the vCloud API
*/

// VApp is a struct representation of application/vnd.vmware.vcloud.vApp+xml
type VApp struct {
	XMLName               xml.Name `xml:"VApp"`
	Text                  string   `xml:",chardata"`
	Xmlns                 string   `xml:"xmlns,attr"`
	Ovf                   string   `xml:"ovf,attr"`
	Vssd                  string   `xml:"vssd,attr"`
	Common                string   `xml:"common,attr"`
	Rasd                  string   `xml:"rasd,attr"`
	Vmw                   string   `xml:"vmw,attr"`
	Vmext                 string   `xml:"vmext,attr"`
	Ovfenv                string   `xml:"ovfenv,attr"`
	Ns9                   string   `xml:"ns9,attr"`
	OvfDescriptorUploaded string   `xml:"ovfDescriptorUploaded,attr"`
	Deployed              string   `xml:"deployed,attr"`
	Status                string   `xml:"status,attr"`
	Name                  string   `xml:"name,attr"`
	ID                    string   `xml:"id,attr"`
	Href                  string   `xml:"href,attr"`
	Type                  string   `xml:"type,attr"`
	Link                  []struct {
		Text string `xml:",chardata"`
		Rel  string `xml:"rel,attr"`
		Href string `xml:"href,attr"`
		Name string `xml:"name,attr"`
		Type string `xml:"type,attr"`
	} `xml:"Link"`
	Description string `xml:"Description"`
	Tasks       struct {
		Text string `xml:",chardata"`
		Task struct {
			Text             string `xml:",chardata"`
			CancelRequested  string `xml:"cancelRequested,attr"`
			ExpiryTime       string `xml:"expiryTime,attr"`
			Operation        string `xml:"operation,attr"`
			OperationName    string `xml:"operationName,attr"`
			ServiceNamespace string `xml:"serviceNamespace,attr"`
			StartTime        string `xml:"startTime,attr"`
			Status           string `xml:"status,attr"`
			Name             string `xml:"name,attr"`
			ID               string `xml:"id,attr"`
			Href             string `xml:"href,attr"`
			Type             string `xml:"type,attr"`
			Owner            struct {
				Text string `xml:",chardata"`
				Href string `xml:"href,attr"`
				ID   string `xml:"id,attr"`
				Name string `xml:"name,attr"`
				Type string `xml:"type,attr"`
			} `xml:"Owner"`
			User struct {
				Text string `xml:",chardata"`
				Href string `xml:"href,attr"`
				ID   string `xml:"id,attr"`
				Name string `xml:"name,attr"`
				Type string `xml:"type,attr"`
			} `xml:"User"`
			Organization struct {
				Text string `xml:",chardata"`
				Href string `xml:"href,attr"`
				ID   string `xml:"id,attr"`
				Name string `xml:"name,attr"`
				Type string `xml:"type,attr"`
			} `xml:"Organization"`
			Progress string `xml:"Progress"`
			Details  string `xml:"Details"`
		} `xml:"Task"`
	} `xml:"Tasks"`
	LeaseSettingsSection struct {
		Text     string `xml:",chardata"`
		Href     string `xml:"href,attr"`
		Type     string `xml:"type,attr"`
		Required string `xml:"required,attr"`
		Info     string `xml:"Info"`
		Link     struct {
			Text string `xml:",chardata"`
			Rel  string `xml:"rel,attr"`
			Href string `xml:"href,attr"`
			Type string `xml:"type,attr"`
		} `xml:"Link"`
		DeploymentLeaseInSeconds string `xml:"DeploymentLeaseInSeconds"`
		StorageLeaseInSeconds    string `xml:"StorageLeaseInSeconds"`
	} `xml:"LeaseSettingsSection"`
	StartupSection struct {
		Text string `xml:",chardata"`
		Ns10 string `xml:"ns10,attr"`
		Type string `xml:"type,attr"`
		Href string `xml:"href,attr"`
		Info string `xml:"Info"`
		Item struct {
			Text        string `xml:",chardata"`
			ID          string `xml:"id,attr"`
			Order       string `xml:"order,attr"`
			StartAction string `xml:"startAction,attr"`
			StartDelay  string `xml:"startDelay,attr"`
			StopAction  string `xml:"stopAction,attr"`
			StopDelay   string `xml:"stopDelay,attr"`
		} `xml:"Item"`
		Link struct {
			Text string `xml:",chardata"`
			Rel  string `xml:"rel,attr"`
			Href string `xml:"href,attr"`
			Type string `xml:"type,attr"`
		} `xml:"Link"`
	} `xml:"StartupSection"`
	NetworkSection struct {
		Text    string `xml:",chardata"`
		Ns10    string `xml:"ns10,attr"`
		Type    string `xml:"type,attr"`
		Href    string `xml:"href,attr"`
		Info    string `xml:"Info"`
		Network struct {
			Text        string `xml:",chardata"`
			Name        string `xml:"name,attr"`
			Description string `xml:"Description"`
		} `xml:"Network"`
	} `xml:"NetworkSection"`
	NetworkConfigSection struct {
		Text          string `xml:",chardata"`
		Href          string `xml:"href,attr"`
		Type          string `xml:"type,attr"`
		Required      string `xml:"required,attr"`
		Info          string `xml:"Info"`
		NetworkConfig struct {
			Text        string `xml:",chardata"`
			NetworkName string `xml:"networkName,attr"`
			Link        struct {
				Text string `xml:",chardata"`
				Rel  string `xml:"rel,attr"`
				Href string `xml:"href,attr"`
			} `xml:"Link"`
			Description   string `xml:"Description"`
			Configuration struct {
				Text     string `xml:",chardata"`
				IPScopes struct {
					Text    string `xml:",chardata"`
					IPScope struct {
						Text               string `xml:",chardata"`
						IsInherited        string `xml:"IsInherited"`
						Gateway            string `xml:"Gateway"`
						Netmask            string `xml:"Netmask"`
						SubnetPrefixLength string `xml:"SubnetPrefixLength"`
						DNS1               string `xml:"Dns1"`
						DNS2               string `xml:"Dns2"`
						IsEnabled          string `xml:"IsEnabled"`
						IPRanges           struct {
							Text    string `xml:",chardata"`
							IPRange struct {
								Text         string `xml:",chardata"`
								StartAddress string `xml:"StartAddress"`
								EndAddress   string `xml:"EndAddress"`
							} `xml:"IpRange"`
						} `xml:"IpRanges"`
					} `xml:"IpScope"`
				} `xml:"IpScopes"`
				ParentNetwork struct {
					Text string `xml:",chardata"`
					Href string `xml:"href,attr"`
					ID   string `xml:"id,attr"`
					Name string `xml:"name,attr"`
				} `xml:"ParentNetwork"`
				FenceMode                      string `xml:"FenceMode"`
				RetainNetInfoAcrossDeployments string `xml:"RetainNetInfoAcrossDeployments"`
				GuestVlanAllowed               string `xml:"GuestVlanAllowed"`
			} `xml:"Configuration"`
			IsDeployed string `xml:"IsDeployed"`
		} `xml:"NetworkConfig"`
	} `xml:"NetworkConfigSection"`
	SnapshotSection struct {
		Text     string `xml:",chardata"`
		Href     string `xml:"href,attr"`
		Type     string `xml:"type,attr"`
		Required string `xml:"required,attr"`
		Info     string `xml:"Info"`
	} `xml:"SnapshotSection"`
	DateCreated string `xml:"DateCreated"`
	Owner       struct {
		Text string `xml:",chardata"`
		Type string `xml:"type,attr"`
		User struct {
			Text string `xml:",chardata"`
			Href string `xml:"href,attr"`
			Name string `xml:"name,attr"`
			Type string `xml:"type,attr"`
		} `xml:"User"`
	} `xml:"Owner"`
	AutoNature        string `xml:"autoNature"`
	InMaintenanceMode string `xml:"InMaintenanceMode"`
	Children          struct {
		Text string `xml:",chardata"`
		VM   struct {
			Text               string `xml:",chardata"`
			NeedsCustomization string `xml:"needsCustomization,attr"`
			Deployed           string `xml:"deployed,attr"`
			Status             string `xml:"status,attr"`
			Name               string `xml:"name,attr"`
			ID                 string `xml:"id,attr"`
			Href               string `xml:"href,attr"`
			Type               string `xml:"type,attr"`
			Link               []struct {
				Text string `xml:",chardata"`
				Rel  string `xml:"rel,attr"`
				Href string `xml:"href,attr"`
				Type string `xml:"type,attr"`
			} `xml:"Link"`
			DateCreated       string `xml:"DateCreated"`
			VAppScopedLocalID string `xml:"VAppScopedLocalId"`
			VMCapabilities    struct {
				Text string `xml:",chardata"`
				Href string `xml:"href,attr"`
				Type string `xml:"type,attr"`
			} `xml:"VmCapabilities"`
			StorageProfile struct {
				Text string `xml:",chardata"`
				Href string `xml:"href,attr"`
				ID   string `xml:"id,attr"`
				Name string `xml:"name,attr"`
				Type string `xml:"type,attr"`
			} `xml:"StorageProfile"`
			VdcComputePolicy struct {
				Text string `xml:",chardata"`
				Href string `xml:"href,attr"`
				ID   string `xml:"id,attr"`
				Name string `xml:"name,attr"`
				Type string `xml:"type,attr"`
			} `xml:"VdcComputePolicy"`
		} `xml:"Vm"`
	} `xml:"Children"`
}

// QueryResultRecords is a struct representation of a query of vApps
type QueryResultRecords struct {
	XMLName  xml.Name `xml:"QueryResultRecords"`
	Text     string   `xml:",chardata"`
	Xmlns    string   `xml:"xmlns,attr"`
	Ovf      string   `xml:"ovf,attr"`
	Vssd     string   `xml:"vssd,attr"`
	Common   string   `xml:"common,attr"`
	Rasd     string   `xml:"rasd,attr"`
	Vmw      string   `xml:"vmw,attr"`
	Vmext    string   `xml:"vmext,attr"`
	Ovfenv   string   `xml:"ovfenv,attr"`
	Ns9      string   `xml:"ns9,attr"`
	Name     string   `xml:"name,attr"`
	Page     string   `xml:"page,attr"`
	PageSize string   `xml:"pageSize,attr"`
	Total    string   `xml:"total,attr"`
	Href     string   `xml:"href,attr"`
	Type     string   `xml:"type,attr"`
	Link     []struct {
		Text string `xml:",chardata"`
		Rel  string `xml:"rel,attr"`
		Href string `xml:"href,attr"`
		Type string `xml:"type,attr"`
	} `xml:"Link"`
	VAppRecord []struct {
		Text                                string `xml:",chardata"`
		CreationDate                        string `xml:"creationDate,attr"`
		IsAutoNature                        string `xml:"isAutoNature,attr"`
		IsBusy                              string `xml:"isBusy,attr"`
		IsDeployed                          string `xml:"isDeployed,attr"`
		IsEnabled                           string `xml:"isEnabled,attr"`
		IsExpired                           string `xml:"isExpired,attr"`
		IsInMaintenanceMode                 string `xml:"isInMaintenanceMode,attr"`
		IsPublic                            string `xml:"isPublic,attr"`
		Name                                string `xml:"name,attr"`
		OwnerName                           string `xml:"ownerName,attr"`
		Snapshot                            string `xml:"snapshot,attr"`
		Status                              string `xml:"status,attr"`
		Vdc                                 string `xml:"vdc,attr"`
		VdcName                             string `xml:"vdcName,attr"`
		Href                                string `xml:"href,attr"`
		TaskStatusName                      string `xml:"taskStatusName,attr"`
		IsAutoDeleteNotified                string `xml:"isAutoDeleteNotified,attr"`
		IsVdcEnabled                        string `xml:"isVdcEnabled,attr"`
		HonorBootOrder                      string `xml:"honorBootOrder,attr"`
		NumberOfVMs                         string `xml:"numberOfVMs,attr"`
		CPUAllocationInMhz                  string `xml:"cpuAllocationInMhz,attr"`
		NumberOfCpus                        string `xml:"numberOfCpus,attr"`
		CPUAllocationMhz                    string `xml:"cpuAllocationMhz,attr"`
		Task                                string `xml:"task,attr"`
		LowestHardwareVersionInVApp         string `xml:"lowestHardwareVersionInVApp,attr"`
		MemoryAllocationMB                  string `xml:"memoryAllocationMB,attr"`
		StorageKB                           string `xml:"storageKB,attr"`
		PvdcHighestSupportedHardwareVersion string `xml:"pvdcHighestSupportedHardwareVersion,attr"`
		IsAutoUndeployNotified              string `xml:"isAutoUndeployNotified,attr"`
		TaskStatus                          string `xml:"taskStatus,attr"`
		TaskDetails                         string `xml:"taskDetails,attr"`
	} `xml:"VAppRecord"`
}

// Task is a struct representation of application/vnd.vmware.vcloud.task+xml
type Task struct {
	XMLName          xml.Name `xml:"Task"`
	Text             string   `xml:",chardata"`
	Xmlns            string   `xml:"xmlns,attr"`
	Ovf              string   `xml:"ovf,attr"`
	Vssd             string   `xml:"vssd,attr"`
	Common           string   `xml:"common,attr"`
	Rasd             string   `xml:"rasd,attr"`
	Vmw              string   `xml:"vmw,attr"`
	Ovfenv           string   `xml:"ovfenv,attr"`
	Vmext            string   `xml:"vmext,attr"`
	Ns9              string   `xml:"ns9,attr"`
	CancelRequested  string   `xml:"cancelRequested,attr"`
	EndTime          string   `xml:"endTime,attr"`
	ExpiryTime       string   `xml:"expiryTime,attr"`
	Operation        string   `xml:"operation,attr"`
	OperationName    string   `xml:"operationName,attr"`
	ServiceNamespace string   `xml:"serviceNamespace,attr"`
	StartTime        string   `xml:"startTime,attr"`
	Status           string   `xml:"status,attr"`
	Name             string   `xml:"name,attr"`
	ID               string   `xml:"id,attr"`
	Href             string   `xml:"href,attr"`
	Type             string   `xml:"type,attr"`
	Owner            struct {
		Text string `xml:",chardata"`
		Href string `xml:"href,attr"`
		ID   string `xml:"id,attr"`
		Name string `xml:"name,attr"`
		Type string `xml:"type,attr"`
	} `xml:"Owner"`
	Error struct {
		Text           string `xml:",chardata"`
		MajorErrorCode string `xml:"majorErrorCode,attr"`
		Message        string `xml:"message,attr"`
		MinorErrorCode string `xml:"minorErrorCode,attr"`
	} `xml:"Error"`
	User struct {
		Text string `xml:",chardata"`
		Href string `xml:"href,attr"`
		ID   string `xml:"id,attr"`
		Name string `xml:"name,attr"`
		Type string `xml:"type,attr"`
	} `xml:"User"`
	Organization struct {
		Text string `xml:",chardata"`
		Href string `xml:"href,attr"`
		ID   string `xml:"id,attr"`
		Name string `xml:"name,attr"`
		Type string `xml:"type,attr"`
	} `xml:"Organization"`
	Progress string `xml:"Progress"`
	Details  string `xml:"Details"`
}

// VcloudError is return when a 400 error code appearss
type VcloudError struct {
	XMLName        xml.Name `xml:"Error"`
	Text           string   `xml:",chardata"`
	Xmlns          string   `xml:"xmlns,attr"`
	Ovf            string   `xml:"ovf,attr"`
	Vssd           string   `xml:"vssd,attr"`
	Common         string   `xml:"common,attr"`
	Rasd           string   `xml:"rasd,attr"`
	Vmw            string   `xml:"vmw,attr"`
	Ovfenv         string   `xml:"ovfenv,attr"`
	Vmext          string   `xml:"vmext,attr"`
	Ns9            string   `xml:"ns9,attr"`
	MajorErrorCode string   `xml:"majorErrorCode,attr"`
	Message        string   `xml:"message,attr"`
	MinorErrorCode string   `xml:"minorErrorCode,attr"`
}

// NetworkConnectionSection represents application/vnd.vmware.vcloud.networkConnectionSection+xml
type NetworkConnectionSection struct {
	XMLName                       xml.Name `xml:"NetworkConnectionSection"`
	Xmlns                         string   `xml:"xmlns,attr"`
	Ovf                           string   `xml:"ovf,attr"`
	Href                          string   `xml:"href,attr"`
	Type                          string   `xml:"type,attr"`
	PrimaryNetworkConnectionIndex int      `xml:"PrimaryNetworkConnectionIndex"`
	NetworkConnection             struct {
		NetworkConnectionIndex  int    `xml:"NetworkConnectionIndex"`
		IPAddress               string `xml:"IpAddress"`
		ExternalIPAddress       string `xml:"ExternalIpAddress"`
		IsConnected             bool   `xml:"IsConnected"`
		MACAddress              string `xml:"MACAddress"`
		IPAddressAllocationMode string `xml:"IpAddressAllocationMode"`
	}
	Link struct {
		Href string `xml:"href,attr"`
		ID   string `xml:"id,attr"`
		Type string `xml:"type,attr"`
		Name string `xml:"name,attr"`
		Rel  string `xml:"rel,attr"`
	}
}
