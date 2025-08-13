package main

// Models for mock Hydra server

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type User struct {
	Username     string   `json:"username"`
	FullName     string   `json:"fullname"`
	EmailAddress string   `json:"emailaddress"`
	Password     string   `json:"password,omitempty"`
	UserRoles    []string `json:"userroles"`
}

type Project struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayname"`
	Owner       string `json:"owner"`
	Description string `json:"description"`
	Homepage    string `json:"homepage"`
	Enabled     bool   `json:"enabled"`
	Hidden      bool   `json:"hidden"`
}

type Jobset struct {
	Name             string                 `json:"name"`
	Project          string                 `json:"project"`
	Description      string                 `json:"description"`
	NixExprInput     *string                `json:"nixexprinput"`
	NixExprPath      *string                `json:"nixexprpath"`
	ErrorMsg         *string                `json:"errormsg"`
	ErrorTime        *int64                 `json:"errortime"`
	LastCheckedTime  *int64                 `json:"lastcheckedtime"`
	TriggerTime      *int64                 `json:"triggertime"`
	Enabled          int                    `json:"enabled"`
	EnableEmail      bool                   `json:"enableemail"`
	Visible          bool                   `json:"visible"`
	EmailOverride    string                 `json:"emailoverride"`
	KeepNr           int                    `json:"keepnr"`
	CheckInterval    int                    `json:"checkinterval"`
	SchedulingShares int                    `json:"schedulingshares"`
	Flake            *string                `json:"flake"`
	Inputs           map[string]JobsetInput `json:"inputs,omitempty"`
}

type JobsetInput struct {
	Name             string `json:"name"`
	Value            string `json:"value"`
	Type             string `json:"type"`
	EmailResponsible bool   `json:"emailresponsible"`
}

type Build struct {
	ID            int     `json:"id"`
	NixName       string  `json:"nixname"`
	Finished      bool    `json:"finished"`
	BuildStatus   *int    `json:"buildstatus"`
	Project       string  `json:"project"`
	Jobset        string  `json:"jobset"`
	Job           string  `json:"job"`
	System        string  `json:"system"`
	StartTime     int64   `json:"starttime"`
	StopTime      int64   `json:"stoptime"`
	Timestamp     int64   `json:"timestamp"`
	JobsetEvals   []int   `json:"jobsetevals"`
	DrvPath       string  `json:"drvpath"`
	BuildProducts []Build `json:"buildproducts"`
}

type JobsetEval struct {
	ID           int    `json:"id"`
	Timestamp    int64  `json:"timestamp"`
	HasNewBuilds bool   `json:"hasnewbuilds"`
	Builds       []int  `json:"builds"`
}

type Evaluations struct {
	First string                   `json:"first"`
	Last  string                   `json:"last"`
	Evals []map[string]*JobsetEval `json:"evals"`
}

type SearchResult struct {
	Projects  []Project `json:"projects"`
	Jobsets   []Jobset  `json:"jobsets"`
	Builds    []Build   `json:"builds"`
	BuildsDrv []Build   `json:"buildsdrv"`
}