package repository

import (
	"time"

	"github.com/chamrasilva89/reservationWeb/internal/models"
)

type DatabaseRepo interface {
	AllUsers() bool

	InsertReservation(res models.Reservation) (int, error)
	InsertRoomRestriction(r models.RoomRestriction) error
	SearchAvailabilityByDatesByRoomID(start, end time.Time, roomID int) (bool, error)
	SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error)

	GetUserByID(id int) (models.User, error)
	UpdateUser(u models.User) error
	Authenticate(email, testPassword string) (int, string, error)

	AllReservations() ([]models.Reservation, error)
	AllNewReservations() ([]models.Reservation, error)
	GetReservationByID(id int) (models.Reservation, error)
	UpdateReservation(u models.Reservation) error
	DeleteReservation(id int) error
	UpdateProcessedForReservation(id, processed int) error

	InsertCustomer(res models.Customer) (int, error)
	AllCustomers() ([]models.Customer, error)
	InsertFile(customerCode, filePath string, customerId int, uniqueFilenameWithExtension string) (int, error)
	GetCustomerByID(id int) (models.Customer, error)
	GetAttachmentsByCustomerID(id int) (models.CustomerImages, error)
	GetTradeShareInforByID(id int) ([]models.TradeLicenseHolder, error)
	GetTradeLicenseInforByID(id int) (models.TradeLicense, error)
	InsertTradeLicense(res models.TradeLicense) (int, error)
	GetMemorandumInforByID(id int) ([]models.Memorandum, error)
	InsertPartner(res models.TradeLicenseHolder) (int, error)
}
