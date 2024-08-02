package tm1go

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ConfigurationService struct to hold any dependencies for configuration operations
type ConfigurationService struct {
	rest *RestService
}

// Configuration struct to hold the configuration settings
type Configuration struct {
	ODataContext                                      string   `json:"@odata.context"`
	ServerName                                        string   `json:"ServerName"`
	AdminHost                                         string   `json:"AdminHost"`
	ProductVersion                                    string   `json:"ProductVersion"`
	PortNumber                                        int      `json:"PortNumber"`
	ClientMessagePortNumber                           int      `json:"ClientMessagePortNumber"`
	HTTPPortNumber                                    int      `json:"HTTPPortNumber"`
	IntegratedSecurityMode                            bool     `json:"IntegratedSecurityMode"`
	SecurityMode                                      string   `json:"SecurityMode"`
	PrincipalName                                     string   `json:"PrincipalName"`
	SecurityPackageName                               string   `json:"SecurityPackageName"`
	ClientCAMURIs                                     []string `json:"ClientCAMURIs"`
	WebCAMURI                                         string   `json:"WebCAMURI"`
	ClientPingCAMPassport                             int      `json:"ClientPingCAMPassport"`
	ServerCAMURI                                      string   `json:"ServerCAMURI"`
	AllowSeparateNandCRules                           bool     `json:"AllowSeparateNandCRules"`
	DistributedOutputDir                              string   `json:"DistributedOutputDir"`
	DisableSandboxing                                 bool     `json:"DisableSandboxing"`
	JobQueuing                                        bool     `json:"JobQueuing"`
	ForceReevaluationOfFeedersForFedCellsOnDataChange bool     `json:"ForceReevaluationOfFeedersForFedCellsOnDataChange"`
	DataBaseDirectory                                 string   `json:"DataBaseDirectory"`
	UnicodeUpperLowerCase                             bool     `json:"UnicodeUpperLowerCase"`
}

type StaticConfiguration struct {
	OdataContext string `json:"@odata.context"`
	ServerName   string `json:"ServerName"`
	Access       struct {
		Network struct {
			IPAddress                        interface{} `json:"IPAddress"`
			IPVersion                        string      `json:"IPVersion"`
			NetRecvBlockingWaitLimit         interface{} `json:"NetRecvBlockingWaitLimit"`
			NetRecvMaxClientIOWaitWithinAPIs interface{} `json:"NetRecvMaxClientIOWaitWithinAPIs"`
			IdleConnectionTimeOut            interface{} `json:"IdleConnectionTimeOut"`
			ReceiveProgressResponseTimeout   interface{} `json:"ReceiveProgressResponseTimeout"`
		} `json:"Network"`
		Authentication struct {
			SecurityPackageName    string      `json:"SecurityPackageName"`
			ServicePrincipalName   interface{} `json:"ServicePrincipalName"`
			IntegratedSecurityMode string      `json:"IntegratedSecurityMode"`
			MaximumLoginAttempts   interface{} `json:"MaximumLoginAttempts"`
		} `json:"Authentication"`
		SSL struct {
			Enable                  bool        `json:"Enable"`
			CertificateID           interface{} `json:"CertificateID"`
			CertAuthority           interface{} `json:"CertAuthority"`
			CertRevocationFile      interface{} `json:"CertRevocationFile"`
			ClientExportServerKeyID interface{} `json:"ClientExportServerKeyID"`
			KeyFile                 interface{} `json:"KeyFile"`
			KeyStashFile            interface{} `json:"KeyStashFile"`
			KeyLabel                interface{} `json:"KeyLabel"`
			TLSCipherList           interface{} `json:"TLSCipherList"`
			FIPSOperationMode       interface{} `json:"FIPSOperationMode"`
			NISTSP800131AMODE       interface{} `json:"NIST_SP800_131A_MODE"`
		} `json:"SSL"`
		CAM struct {
			CAMUseSSL                 interface{} `json:"CAMUseSSL"`
			ClientURI                 interface{} `json:"ClientURI"`
			ServerURIs                interface{} `json:"ServerURIs"`
			PortalVariableFile        interface{} `json:"PortalVariableFile"`
			ClientPingCAMPassport     interface{} `json:"ClientPingCAMPassport"`
			ServerCAMURIRetryAttempts interface{} `json:"ServerCAMURIRetryAttempts"`
			CreateNewCAMClients       interface{} `json:"CreateNewCAMClients"`
		} `json:"CAM"`
		LDAP struct {
			Enable                  bool        `json:"Enable"`
			Host                    interface{} `json:"Host"`
			Port                    interface{} `json:"Port"`
			UseServerAccount        interface{} `json:"UseServerAccount"`
			VerifyCertServerName    interface{} `json:"VerifyCertServerName"`
			VerifyServerSSLCert     interface{} `json:"VerifyServerSSLCert"`
			SkipSSLCertVerification interface{} `json:"SkipSSLCertVerification"`
			SkipSSLCRLVerification  interface{} `json:"SkipSSLCRLVerification"`
			WellKnownUserName       interface{} `json:"WellKnownUserName"`
			PasswordFile            interface{} `json:"PasswordFile"`
			PasswordKeyFile         interface{} `json:"PasswordKeyFile"`
			SearchBase              interface{} `json:"SearchBase"`
			SearchField             interface{} `json:"SearchField"`
		} `json:"LDAP"`
		CAPI struct {
			Port                   int         `json:"Port"`
			ClientMessagePort      int         `json:"ClientMessagePort"`
			MessageCompression     interface{} `json:"MessageCompression"`
			ProgressMessage        bool        `json:"ProgressMessage"`
			ClientVersionMaximum   interface{} `json:"ClientVersionMaximum"`
			ClientVersionMinimum   interface{} `json:"ClientVersionMinimum"`
			ClientVersionPrecision interface{} `json:"ClientVersionPrecision"`
		} `json:"CAPI"`
		HTTP struct {
			Port                     int           `json:"Port"`
			SessionTimeout           interface{}   `json:"SessionTimeout"`
			SessionMaxRequests       interface{}   `json:"SessionMaxRequests"`
			RequestEntityMaxSizeInKB interface{}   `json:"RequestEntityMaxSizeInKB"`
			OriginAllowList          []interface{} `json:"OriginAllowList"`
		} `json:"HTTP"`
	} `json:"Access"`
	Administration struct {
		ServerName                   string      `json:"ServerName"`
		AdminHost                    string      `json:"AdminHost"`
		Language                     interface{} `json:"Language"`
		DataBaseDirectory            string      `json:"DataBaseDirectory"`
		UnicodeUpperLowerCase        interface{} `json:"UnicodeUpperLowerCase"`
		MaskUserNameInServerTools    interface{} `json:"MaskUserNameInServerTools"`
		AllowReadOnlyChoreReschedule interface{} `json:"AllowReadOnlyChoreReschedule"`
		DisableSandboxing            interface{} `json:"DisableSandboxing"`
		RunningInBackground          interface{} `json:"RunningInBackground"`
		StartupChores                interface{} `json:"StartupChores"`
		PerformanceMonitorOn         interface{} `json:"PerformanceMonitorOn"`
		PerfMonActive                interface{} `json:"PerfMonActive"`
		EnableSandboxDimension       interface{} `json:"EnableSandboxDimension"`
		Clients                      struct {
			PasswordMinimumLength        interface{} `json:"PasswordMinimumLength"`
			ClientPropertiesSyncInterval interface{} `json:"ClientPropertiesSyncInterval"`
			RetainNonCAMGroupMembership  interface{} `json:"RetainNonCAMGroupMembership"`
		} `json:"Clients"`
		AuditLog struct {
			Enable                  bool        `json:"Enable"`
			UpdateInterval          string      `json:"UpdateInterval"`
			MaxFileSizeKilobytes    int         `json:"MaxFileSizeKilobytes"`
			MaxQueryMemoryKilobytes interface{} `json:"MaxQueryMemoryKilobytes"`
		} `json:"AuditLog"`
		DebugLog struct {
			LoggingDirectory interface{} `json:"LoggingDirectory"`
		} `json:"DebugLog"`
		ServerLog struct {
			Enable              bool        `json:"Enable"`
			LogReleaseLineCount interface{} `json:"LogReleaseLineCount"`
		} `json:"ServerLog"`
		EventLog struct {
			Enable                           interface{} `json:"Enable"`
			ScanFrequency                    interface{} `json:"ScanFrequency"`
			ThresholdForThreadRunningTime    interface{} `json:"ThresholdForThreadRunningTime"`
			ThresholdForThreadWaitingTime    interface{} `json:"ThresholdForThreadWaitingTime"`
			ThresholdForThreadBlockingNumber interface{} `json:"ThresholdForThreadBlockingNumber"`
			ThresholdForPooledMemoryInMB     interface{} `json:"ThresholdForPooledMemoryInMB"`
		} `json:"EventLog"`
		TopLog struct {
			Enable        interface{} `json:"Enable"`
			ScanMode      interface{} `json:"ScanMode"`
			ScanFrequency interface{} `json:"ScanFrequency"`
		} `json:"TopLog"`
		Java struct {
			ClassPath interface{} `json:"ClassPath"`
			JVMPath   interface{} `json:"JVMPath"`
			JVMArgs   interface{} `json:"JVMArgs"`
		} `json:"Java"`
		ExternalDatabase struct {
			OracleErrorForceRowStatus interface{} `json:"OracleErrorForceRowStatus"`
			SQLFetchType              interface{} `json:"SQLFetchType"`
			SQLRowsetSize             interface{} `json:"SQLRowsetSize"`
			ODBCLibraryPath           interface{} `json:"ODBCLibraryPath"`
			TM1ConnectorforSAP        interface{} `json:"TM1ConnectorforSAP"`
			UseNewConnectorforSAP     interface{} `json:"UseNewConnectorforSAP"`
			ODBCTimeoutInSeconds      interface{} `json:"ODBCTimeoutInSeconds"`
		} `json:"ExternalDatabase"`
		TM1Web struct {
			ExcelWebPublishEnabled interface{} `json:"ExcelWebPublishEnabled"`
		} `json:"TM1Web"`
		FileRetry struct {
			FileRetryCount             interface{} `json:"FileRetryCount"`
			FileRetryDelayMilliseconds interface{} `json:"FileRetryDelayMilliseconds"`
			FileRetryFileSpec          interface{} `json:"FileRetryFileSpec"`
		} `json:"FileRetry"`
		DownTime string `json:"DownTime"`
	} `json:"Administration"`
	Modelling struct {
		MDXSelectCalculatedMemberInputs interface{} `json:"MDXSelectCalculatedMemberInputs"`
		DefaultMeasuresDimension        interface{} `json:"DefaultMeasuresDimension"`
		UserDefinedCalculations         interface{} `json:"UserDefinedCalculations"`
		EnableNewHierarchyCreation      bool        `json:"EnableNewHierarchyCreation"`
		Spreading                       struct {
			SpreadingPrecision          interface{} `json:"SpreadingPrecision"`
			ProportionSpreadToZeroCells interface{} `json:"ProportionSpreadToZeroCells"`
		} `json:"Spreading"`
		TI struct {
			CognosTM1InterfacePath interface{} `json:"CognosTM1InterfacePath"`
			UseExcelSerialDate     interface{} `json:"UseExcelSerialDate"`
			MaximumTILockObjects   interface{} `json:"MaximumTILockObjects"`
			EnableTIDebugging      bool        `json:"EnableTIDebugging"`
		} `json:"TI"`
		Rules struct {
			AllowSeparateNandCRules                           bool        `json:"AllowSeparateNandCRules"`
			AutomaticallyAddCubeDependencies                  interface{} `json:"AutomaticallyAddCubeDependencies"`
			RulesOverwriteCellsOnLoad                         interface{} `json:"RulesOverwriteCellsOnLoad"`
			ForceReevaluationOfFeedersForFedCellsOnDataChange bool        `json:"ForceReevaluationOfFeedersForFedCellsOnDataChange"`
		} `json:"Rules"`
		Startup struct {
			PersistentFeeders           bool        `json:"PersistentFeeders"`
			SkipLoadingAliases          interface{} `json:"SkipLoadingAliases"`
			MaximumCubeLoadThreads      interface{} `json:"MaximumCubeLoadThreads"`
			LoadPrivateSubsetsOnStartup interface{} `json:"LoadPrivateSubsetsOnStartup"`
		} `json:"Startup"`
		Synchronization struct {
			SyncUnitSize         interface{} `json:"SyncUnitSize"`
			MaximumSynchAttempts interface{} `json:"MaximumSynchAttempts"`
		} `json:"Synchronization"`
	} `json:"Modelling"`
	Performance struct {
		PrivilegeGenerationOptimization interface{} `json:"PrivilegeGenerationOptimization"`
		Memory                          struct {
			ApplyMaximumViewSizeToEntireTransaction interface{} `json:"ApplyMaximumViewSizeToEntireTransaction"`
			DisableMemoryCache                      interface{} `json:"DisableMemoryCache"`
			CacheFriendlyMalloc                     interface{} `json:"CacheFriendlyMalloc"`
			MaximumViewSizeMB                       interface{} `json:"MaximumViewSizeMB"`
			MaximumUserSandboxSizeMB                interface{} `json:"MaximumUserSandboxSizeMB"`
			MaximumMemoryForSubsetUndoKB            interface{} `json:"MaximumMemoryForSubsetUndoKB"`
			LockPagesInMemory                       interface{} `json:"LockPagesInMemory"`
		} `json:"Memory"`
		MTCubeLoad struct {
			Enabled          interface{} `json:"Enabled"`
			Weight           interface{} `json:"Weight"`
			MinFileSize      interface{} `json:"MinFileSize"`
			UseBookmarkFiles interface{} `json:"UseBookmarkFiles"`
		} `json:"MTCubeLoad"`
		MTFeeders struct {
			Enabled   interface{} `json:"Enabled"`
			AtStartup interface{} `json:"AtStartup"`
		} `json:"MTFeeders"`
		MTQ struct {
			UseAllThreads                      interface{} `json:"UseAllThreads"`
			NumberOfThreadsToUse               interface{} `json:"NumberOfThreadsToUse"`
			SingleCellConsolidation            interface{} `json:"SingleCellConsolidation"`
			ImmediateCheckForSplit             interface{} `json:"ImmediateCheckForSplit"`
			OperationProgressCheckSkipLoopSize interface{} `json:"OperationProgressCheckSkipLoopSize"`
			MTFeeders                          interface{} `json:"MTFeeders"`
			MTFeedersAtStartup                 interface{} `json:"MTFeedersAtStartup"`
			MTQQuery                           interface{} `json:"MTQQuery"`
		} `json:"MTQ"`
		Locking struct {
			SubsetElementLockBreathing            interface{} `json:"SubsetElementLockBreathing"`
			UseLocalCopiesForPublicDynamicSubsets interface{} `json:"UseLocalCopiesForPublicDynamicSubsets"`
			PullInvalidationSubsets               interface{} `json:"PullInvalidationSubsets"`
		} `json:"Locking"`
		ViewCalculation struct {
			MagnitudeDifferenceToBeZero         interface{} `json:"MagnitudeDifferenceToBeZero"`
			CheckFeedersMaximumCells            interface{} `json:"CheckFeedersMaximumCells"`
			CalculationThresholdForStorage      interface{} `json:"CalculationThresholdForStorage"`
			ViewConsolidationOptimization       interface{} `json:"ViewConsolidationOptimization"`
			ViewConsolidationOptimizationMethod interface{} `json:"ViewConsolidationOptimizationMethod"`
		} `json:"ViewCalculation"`
		Stargate struct {
			ZeroWeightOptimization          interface{} `json:"ZeroWeightOptimization"`
			AllRuleCalcStargateOptimization interface{} `json:"AllRuleCalcStargateOptimization"`
			UseStargateForRules             interface{} `json:"UseStargateForRules"`
		} `json:"Stargate"`
		JobQueuing struct {
			Enable          interface{} `json:"Enable"`
			ThreadSleepTime interface{} `json:"ThreadSleepTime"`
			ThreadPoolSize  interface{} `json:"ThreadPoolSize"`
			MaxWaitTime     interface{} `json:"MaxWaitTime"`
		} `json:"JobQueuing"`
	} `json:"Performance"`
}

type ActiveConfiguration struct {
	OdataContext string `json:"@odata.context"`
	ServerName   string `json:"ServerName"`
	Access       struct {
		Network struct {
			IPAddress                        interface{} `json:"IPAddress"`
			IPVersion                        string      `json:"IPVersion"`
			NetRecvBlockingWaitLimit         string      `json:"NetRecvBlockingWaitLimit"`
			NetRecvMaxClientIOWaitWithinAPIs string      `json:"NetRecvMaxClientIOWaitWithinAPIs"`
			IdleConnectionTimeOut            string      `json:"IdleConnectionTimeOut"`
			ReceiveProgressResponseTimeout   string      `json:"ReceiveProgressResponseTimeout"`
		} `json:"Network"`
		Authentication struct {
			SecurityPackageName    string `json:"SecurityPackageName"`
			ServicePrincipalName   string `json:"ServicePrincipalName"`
			IntegratedSecurityMode string `json:"IntegratedSecurityMode"`
			MaximumLoginAttempts   int    `json:"MaximumLoginAttempts"`
		} `json:"Authentication"`
		SSL struct {
			Enable                  bool        `json:"Enable"`
			CertificateID           interface{} `json:"CertificateID"`
			CertAuthority           interface{} `json:"CertAuthority"`
			CertRevocationFile      interface{} `json:"CertRevocationFile"`
			ClientExportServerKeyID interface{} `json:"ClientExportServerKeyID"`
			KeyFile                 string      `json:"KeyFile"`
			KeyStashFile            string      `json:"KeyStashFile"`
			KeyLabel                interface{} `json:"KeyLabel"`
			TLSCipherList           interface{} `json:"TLSCipherList"`
			FIPSOperationMode       string      `json:"FIPSOperationMode"`
			NISTSP800131AMODE       bool        `json:"NIST_SP800_131A_MODE"`
		} `json:"SSL"`
		CAM struct {
			CAMUseSSL                 bool          `json:"CAMUseSSL"`
			ClientURI                 interface{}   `json:"ClientURI"`
			ServerURIs                []interface{} `json:"ServerURIs"`
			PortalVariableFile        interface{}   `json:"PortalVariableFile"`
			ClientPingCAMPassport     string        `json:"ClientPingCAMPassport"`
			ServerCAMURIRetryAttempts int           `json:"ServerCAMURIRetryAttempts"`
			CreateNewCAMClients       bool          `json:"CreateNewCAMClients"`
		} `json:"CAM"`
		LDAP struct {
			Enable                  bool          `json:"Enable"`
			Host                    interface{}   `json:"Host"`
			Port                    int           `json:"Port"`
			UseServerAccount        bool          `json:"UseServerAccount"`
			VerifyCertServerName    []interface{} `json:"VerifyCertServerName"`
			VerifyServerSSLCert     bool          `json:"VerifyServerSSLCert"`
			SkipSSLCertVerification bool          `json:"SkipSSLCertVerification"`
			SkipSSLCRLVerification  bool          `json:"SkipSSLCRLVerification"`
			WellKnownUserName       interface{}   `json:"WellKnownUserName"`
			PasswordFile            interface{}   `json:"PasswordFile"`
			PasswordKeyFile         interface{}   `json:"PasswordKeyFile"`
			SearchBase              interface{}   `json:"SearchBase"`
			SearchField             string        `json:"SearchField"`
		} `json:"LDAP"`
		CAPI struct {
			Port                   int  `json:"Port"`
			ClientMessagePort      int  `json:"ClientMessagePort"`
			MessageCompression     bool `json:"MessageCompression"`
			ProgressMessage        bool `json:"ProgressMessage"`
			ClientVersionMaximum   int  `json:"ClientVersionMaximum"`
			ClientVersionMinimum   int  `json:"ClientVersionMinimum"`
			ClientVersionPrecision int  `json:"ClientVersionPrecision"`
		} `json:"CAPI"`
		HTTP struct {
			Port                     int           `json:"Port"`
			SessionTimeout           string        `json:"SessionTimeout"`
			SessionMaxRequests       int           `json:"SessionMaxRequests"`
			RequestEntityMaxSizeInKB int           `json:"RequestEntityMaxSizeInKB"`
			OriginAllowList          []interface{} `json:"OriginAllowList"`
		} `json:"HTTP"`
	} `json:"Access"`
	Administration struct {
		ServerName                   string      `json:"ServerName"`
		AdminHost                    interface{} `json:"AdminHost"`
		Language                     interface{} `json:"Language"`
		DataBaseDirectory            string      `json:"DataBaseDirectory"`
		UnicodeUpperLowerCase        bool        `json:"UnicodeUpperLowerCase"`
		MaskUserNameInServerTools    bool        `json:"MaskUserNameInServerTools"`
		AllowReadOnlyChoreReschedule bool        `json:"AllowReadOnlyChoreReschedule"`
		DisableSandboxing            bool        `json:"DisableSandboxing"`
		RunningInBackground          interface{} `json:"RunningInBackground"`
		StartupChores                interface{} `json:"StartupChores"`
		PerformanceMonitorOn         bool        `json:"PerformanceMonitorOn"`
		PerfMonActive                bool        `json:"PerfMonActive"`
		EnableSandboxDimension       bool        `json:"EnableSandboxDimension"`
		Clients                      struct {
			PasswordMinimumLength        int         `json:"PasswordMinimumLength"`
			ClientPropertiesSyncInterval string      `json:"ClientPropertiesSyncInterval"`
			RetainNonCAMGroupMembership  interface{} `json:"RetainNonCAMGroupMembership"`
		} `json:"Clients"`
		AuditLog struct {
			Enable                  bool   `json:"Enable"`
			UpdateInterval          string `json:"UpdateInterval"`
			MaxFileSizeKilobytes    int    `json:"MaxFileSizeKilobytes"`
			MaxQueryMemoryKilobytes int    `json:"MaxQueryMemoryKilobytes"`
		} `json:"AuditLog"`
		DebugLog struct {
			LoggingDirectory string `json:"LoggingDirectory"`
		} `json:"DebugLog"`
		ServerLog struct {
			Enable              bool `json:"Enable"`
			LogReleaseLineCount int  `json:"LogReleaseLineCount"`
		} `json:"ServerLog"`
		EventLog struct {
			Enable                           bool   `json:"Enable"`
			ScanFrequency                    string `json:"ScanFrequency"`
			ThresholdForThreadRunningTime    string `json:"ThresholdForThreadRunningTime"`
			ThresholdForThreadWaitingTime    string `json:"ThresholdForThreadWaitingTime"`
			ThresholdForThreadBlockingNumber int    `json:"ThresholdForThreadBlockingNumber"`
			ThresholdForPooledMemoryInMB     int    `json:"ThresholdForPooledMemoryInMB"`
		} `json:"EventLog"`
		TopLog struct {
			Enable        bool   `json:"Enable"`
			ScanMode      string `json:"ScanMode"`
			ScanFrequency string `json:"ScanFrequency"`
		} `json:"TopLog"`
		Java struct {
			ClassPath interface{} `json:"ClassPath"`
			JVMPath   interface{} `json:"JVMPath"`
			JVMArgs   interface{} `json:"JVMArgs"`
		} `json:"Java"`
		ExternalDatabase struct {
			OracleErrorForceRowStatus interface{} `json:"OracleErrorForceRowStatus"`
			SQLFetchType              interface{} `json:"SQLFetchType"`
			SQLRowsetSize             int         `json:"SQLRowsetSize"`
			ODBCLibraryPath           interface{} `json:"ODBCLibraryPath"`
			TM1ConnectorforSAP        bool        `json:"TM1ConnectorforSAP"`
			UseNewConnectorforSAP     bool        `json:"UseNewConnectorforSAP"`
			ODBCTimeoutInSeconds      int         `json:"ODBCTimeoutInSeconds"`
		} `json:"ExternalDatabase"`
		TM1Web struct {
			ExcelWebPublishEnabled bool `json:"ExcelWebPublishEnabled"`
		} `json:"TM1Web"`
		FileRetry struct {
			FileRetryCount             int      `json:"FileRetryCount"`
			FileRetryDelayMilliseconds int      `json:"FileRetryDelayMilliseconds"`
			FileRetryFileSpec          []string `json:"FileRetryFileSpec"`
		} `json:"FileRetry"`
		DownTime string `json:"DownTime"`
	} `json:"Administration"`
	Modelling struct {
		MDXSelectCalculatedMemberInputs bool `json:"MDXSelectCalculatedMemberInputs"`
		DefaultMeasuresDimension        bool `json:"DefaultMeasuresDimension"`
		UserDefinedCalculations         bool `json:"UserDefinedCalculations"`
		EnableNewHierarchyCreation      bool `json:"EnableNewHierarchyCreation"`
		Spreading                       struct {
			SpreadingPrecision          float64 `json:"SpreadingPrecision"`
			ProportionSpreadToZeroCells bool    `json:"ProportionSpreadToZeroCells"`
		} `json:"Spreading"`
		TI struct {
			CognosTM1InterfacePath interface{} `json:"CognosTM1InterfacePath"`
			UseExcelSerialDate     bool        `json:"UseExcelSerialDate"`
			MaximumTILockObjects   int         `json:"MaximumTILockObjects"`
			EnableTIDebugging      bool        `json:"EnableTIDebugging"`
		} `json:"TI"`
		Rules struct {
			AllowSeparateNandCRules                           bool `json:"AllowSeparateNandCRules"`
			AutomaticallyAddCubeDependencies                  bool `json:"AutomaticallyAddCubeDependencies"`
			RulesOverwriteCellsOnLoad                         bool `json:"RulesOverwriteCellsOnLoad"`
			ForceReevaluationOfFeedersForFedCellsOnDataChange bool `json:"ForceReevaluationOfFeedersForFedCellsOnDataChange"`
		} `json:"Rules"`
		Startup struct {
			PersistentFeeders           bool        `json:"PersistentFeeders"`
			SkipLoadingAliases          interface{} `json:"SkipLoadingAliases"`
			MaximumCubeLoadThreads      int         `json:"MaximumCubeLoadThreads"`
			LoadPrivateSubsetsOnStartup bool        `json:"LoadPrivateSubsetsOnStartup"`
		} `json:"Startup"`
		Synchronization struct {
			SyncUnitSize         int `json:"SyncUnitSize"`
			MaximumSynchAttempts int `json:"MaximumSynchAttempts"`
		} `json:"Synchronization"`
	} `json:"Modelling"`
	Performance struct {
		PrivilegeGenerationOptimization bool `json:"PrivilegeGenerationOptimization"`
		Memory                          struct {
			ApplyMaximumViewSizeToEntireTransaction bool `json:"ApplyMaximumViewSizeToEntireTransaction"`
			DisableMemoryCache                      bool `json:"DisableMemoryCache"`
			CacheFriendlyMalloc                     bool `json:"CacheFriendlyMalloc"`
			MaximumViewSizeMB                       int  `json:"MaximumViewSizeMB"`
			MaximumUserSandboxSizeMB                int  `json:"MaximumUserSandboxSizeMB"`
			MaximumMemoryForSubsetUndoKB            int  `json:"MaximumMemoryForSubsetUndoKB"`
			LockPagesInMemory                       bool `json:"LockPagesInMemory"`
		} `json:"Memory"`
		MTCubeLoad struct {
			Enabled          bool `json:"Enabled"`
			Weight           int  `json:"Weight"`
			MinFileSize      int  `json:"MinFileSize"`
			UseBookmarkFiles bool `json:"UseBookmarkFiles"`
		} `json:"MTCubeLoad"`
		MTFeeders struct {
			Enabled   bool `json:"Enabled"`
			AtStartup bool `json:"AtStartup"`
		} `json:"MTFeeders"`
		MTQ struct {
			UseAllThreads                      bool `json:"UseAllThreads"`
			NumberOfThreadsToUse               int  `json:"NumberOfThreadsToUse"`
			SingleCellConsolidation            bool `json:"SingleCellConsolidation"`
			ImmediateCheckForSplit             bool `json:"ImmediateCheckForSplit"`
			OperationProgressCheckSkipLoopSize int  `json:"OperationProgressCheckSkipLoopSize"`
			MTFeeders                          bool `json:"MTFeeders"`
			MTFeedersAtStartup                 bool `json:"MTFeedersAtStartup"`
			MTQQuery                           bool `json:"MTQQuery"`
		} `json:"MTQ"`
		Locking struct {
			SubsetElementLockBreathing            bool `json:"SubsetElementLockBreathing"`
			UseLocalCopiesForPublicDynamicSubsets bool `json:"UseLocalCopiesForPublicDynamicSubsets"`
			PullInvalidationSubsets               bool `json:"PullInvalidationSubsets"`
		} `json:"Locking"`
		ViewCalculation struct {
			MagnitudeDifferenceToBeZero         int    `json:"MagnitudeDifferenceToBeZero"`
			CheckFeedersMaximumCells            int    `json:"CheckFeedersMaximumCells"`
			CalculationThresholdForStorage      int    `json:"CalculationThresholdForStorage"`
			ViewConsolidationOptimization       bool   `json:"ViewConsolidationOptimization"`
			ViewConsolidationOptimizationMethod string `json:"ViewConsolidationOptimizationMethod"`
		} `json:"ViewCalculation"`
		Stargate struct {
			ZeroWeightOptimization          bool `json:"ZeroWeightOptimization"`
			AllRuleCalcStargateOptimization bool `json:"AllRuleCalcStargateOptimization"`
			UseStargateForRules             bool `json:"UseStargateForRules"`
		} `json:"Stargate"`
		JobQueuing struct {
			Enable          bool   `json:"Enable"`
			ThreadSleepTime string `json:"ThreadSleepTime"`
			ThreadPoolSize  int    `json:"ThreadPoolSize"`
			MaxWaitTime     string `json:"MaxWaitTime"`
		} `json:"JobQueuing"`
	} `json:"Performance"`
}

// NewConfigurationService creates a new ConfigurationService with the provided RestClient
func NewConfigurationService(rest *RestService) *ConfigurationService {
	return &ConfigurationService{rest: rest}
}

// GetAll retrieves all configuration settings, removing the @odata context entry
func (cs *ConfigurationService) GetAll() (Configuration, error) {
	configuration := Configuration{}
	url := "/Configuration"
	resp, err := cs.rest.GET(url, nil, 0, nil)
	if err != nil {
		return configuration, err
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&configuration)
	if err != nil {
		return configuration, err
	}

	return configuration, nil
}

// GetServerName asks the TM1 Server for its name
func (cs *ConfigurationService) GetServerName() (string, error) {
	url := "/Configuration/ServerName/$value"
	resp, err := cs.rest.GET(url, nil, 0, nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	return string(body), err
}

// GetProductVersion asks the TM1 Server for its version.
func (cs *ConfigurationService) GetProductVersion() (string, error) {
	url := "/Configuration/ProductVersion/$value"
	resp, err := cs.rest.GET(url, nil, 0, nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	return string(body), err
}

// GetAdminHost is deprecated in version 12.0.0
func (cs *ConfigurationService) GetAdminHost() (string, error) {
	if IsV1GreaterOrEqualToV2(cs.rest.version, "12.0.0") {
		err := fmt.Errorf("GetDataDirectory is deprecated in version 12.0.0")
		return "", err
	}
	url := "/Configuration/AdminHost/$value"
	resp, err := cs.rest.GET(url, nil, 0, nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	return string(body), err
}

// GetDataDirectory is deprecated in version 12.0.0
func (cs *ConfigurationService) GetDataDirectory() (string, error) {
	if IsV1GreaterOrEqualToV2(cs.rest.version, "12.0.0") {
		err := fmt.Errorf("GetDataDirectory is deprecated in version 12.0.0")
		return "", err
	}

	url := "/Configuration/DataBaseDirectory/$value"
	resp, err := cs.rest.GET(url, nil, 0, nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	return string(body), err
}

// GetStatic reads TM1 config settings as dictionary from TM1 Server.
// Requires Ops Admin privilege.
func (cs *ConfigurationService) GetStatic() (StaticConfiguration, error) {

	staticConfiguration := StaticConfiguration{}
	url := "/StaticConfiguration"
	resp, err := cs.rest.GET(url, nil, 0, nil)
	if err != nil {
		return staticConfiguration, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&staticConfiguration)
	return staticConfiguration, err
}

// GetActive reads effective TM1 config settings as dictionary from TM1 Server.
// Requires Ops Admin privilege.
func (cs *ConfigurationService) GetActive() (ActiveConfiguration, error) {
	if !cs.rest.IsOpsAdmin() {
		return ActiveConfiguration{}, fmt.Errorf("GetActive requires Ops Admin privilege")
	}
	activeConfiguration := ActiveConfiguration{}
	url := "/ActiveConfiguration"
	resp, err := cs.rest.GET(url, nil, 0, nil)
	if err != nil {
		return activeConfiguration, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&activeConfiguration)
	if err != nil {
		return activeConfiguration, err
	}

	return activeConfiguration, err
}

// UpdateStatic updates the .cfg file and triggers TM1 to re-read the file.
// Requires Ops Admin privilege.
func (cs *ConfigurationService) UpdateStatic(configuration map[string]interface{}) (*http.Response, error) {
	if !cs.rest.IsOpsAdmin() {
		return nil, fmt.Errorf("UpdateStatic requires Ops Admin privilege")
	}
	url := "/StaticConfiguration"
	// convert configuration to json string
	jsonStr, err := json.Marshal(configuration)
	if err != nil {
		return nil, err
	}

	return cs.rest.PATCH(url, string(jsonStr), nil, 0, nil)
}
