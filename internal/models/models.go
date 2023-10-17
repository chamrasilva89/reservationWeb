package models

import "time"

type User struct {
	ID          int
	FirstName   string
	LastName    string
	Email       string
	Password    string
	AccessLevel int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Room struct {
	ID         int
	RoomName   string
	CreatedAt  time.Time
	UppdatedAt time.Time
}

type Restriction struct {
	ID              int
	RestrictionName string
	CreatedAt       time.Time
	UppdatedAt      time.Time
}

type Reservation struct {
	ID        int
	FirstName string
	LastName  string
	Email     string
	Phone     string
	StartDate time.Time
	EndDate   time.Time
	RoomID    int
	CreatedAt time.Time
	UpdatedAt time.Time
	Room      Room
	Processed int
}

type RoomRestriction struct {
	ID            int
	StartDate     time.Time
	EndDate       time.Time
	RoomID        int
	ReservationID int
	RestrictionID int
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Room          Room
	Reservation   Reservation
	Restriction   Restriction
}

type Customer struct {
	CustomerId       int
	CustomerCode     string
	CustomerName     string
	ContactNo        string
	ContactPerson    string
	Email            string
	MobileNo         string
	BusinessName     string
	Status           string
	LocationDetails  string
	NatureOfBusiness string
	MarketedBy       string
	MarketerName     string
	MarketerEmail    string
	Photos           []string // You can use a slice of strings to store photo filenames
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type CustomerImages struct {
	CustomerId   int
	CustomerCode string
	Attachments  []Attachment
}

type Attachment struct {
	File_id  int
	FilePath string
	FileName string // You may want to add a field for the file name as well

}

type TradeLicenseHolder struct {
	TradeLicenseID  int
	CustomerId      int
	CustomerCode    string
	CustomerName    string
	ShareHolderID   int
	ShareHolderName string
	ShareHolderRole string
	ShNationality   string
	ShNoOfShares    int
	ShEmirateID     string
	ShEmIDExp       time.Time
	ShPassport      string
	ShPassportExp   time.Time
	ShIDFilepath    string
	ShIDFileName    string
	ShPassFilepath  string
	ShPassFileName  string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type TradeLicense struct {
	TradeLicenseID   int
	TradeLicenseNo   string
	CustomerId       int
	CustomerCode     string
	Emirate          string
	MohreNo          string
	TradeName        string
	LegalStatus      string
	EstablishDate    time.Time
	RegistrationDate time.Time
	LicenseExpiry    time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
	FilePath         string
	FileName         string
}

type Memorandum struct {
	TradeLicenseID     int
	MemorandumID       int
	CustomerId         int
	CustomerCode       string
	RepresentativeName string
	RepNoOfShares      string
	RepEmID            string
	RepEmIDExp         time.Time
	RepPassport        string
	RepPassportExp     time.Time
	RepIDFilepath      string
	RepPassFilepath    string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
