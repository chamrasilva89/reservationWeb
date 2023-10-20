package dbrepo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/chamrasilva89/reservationWeb/internal/models"
	"golang.org/x/crypto/bcrypt"
)

func (m *postgresDBRepo) AllUsers() bool {
	return true
}

func (m *postgresDBRepo) InsertReservation(res models.Reservation) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	//
	var newID int
	stmt := `insert into reservations (first_name,last_name,email,phone,start_date,
					end_date,room_id,created_at,updated_at) values($1,$2,$3,$4,$5,$6,$7,$8,$9) returning id`

	err := m.DB.QueryRowContext(ctx, stmt,
		res.FirstName,
		res.LastName,
		res.Email,
		res.Phone,
		res.StartDate,
		res.EndDate,
		res.RoomID,
		time.Now(),
		time.Now(),
	).Scan(&newID)

	if err != nil {
		return newID, err
	}

	return 0, nil
}

func (m *postgresDBRepo) InsertRoomRestriction(r models.RoomRestriction) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stmt := `INSERT INTO public.room_restrictions (start_date, end_date, room_id, reservation_id, created_at, updated_at,restriction_id)
		values ($1,$2,$3,$4,$5,$6,$7)`

	_, err := m.DB.ExecContext(ctx, stmt,
		r.StartDate,
		r.EndDate,
		r.RoomID,
		r.ReservationID,
		time.Now(),
		time.Now(),
		r.RestrictionID,
	)
	if err != nil {
		return err
	}
	return nil
}

func (m *postgresDBRepo) SearchAvailabilityByDatesByRoomID(start, end time.Time, roomID int) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var numRows int
	query := `select count(id) from room_restrictions where $1 < end_date $2 > start_date`

	row := m.DB.QueryRowContext(ctx, query, start, end)
	err := row.Scan(&numRows)
	if err != nil {
		return false, err
	}
	if numRows == 0 {
		return true, nil
	}
	return false, nil

}

// SearchAvailabilityForAllRooms returns a slice of available rooms, if any, for given date range
func (m *postgresDBRepo) SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var rooms []models.Room

	query := `
		select
			r.id, r.room_name
		from
			rooms r
		where r.id not in 
		(select room_id from room_restrictions rr where $1 < rr.end_date and $2 > rr.start_date);
		`

	rows, err := m.DB.QueryContext(ctx, query, start, end)
	if err != nil {
		return rooms, err
	}

	for rows.Next() {
		var room models.Room
		err := rows.Scan(
			&room.ID,
			&room.RoomName,
		)
		if err != nil {
			return rooms, err
		}
		rooms = append(rooms, room)
	}

	if err = rows.Err(); err != nil {
		return rooms, err
	}

	return rooms, nil
}

func (m *postgresDBRepo) GetUserByID(id int) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `select id,first_name,last_name,email,password,access_level, created_at,updated_at
	from users where id = $1`

	row := m.DB.QueryRowContext(ctx, query, id)
	var u models.User
	err := row.Scan(
		&u.ID, &u.FirstName, &u.LastName, &u.Email, &u.Password, &u.AccessLevel, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return u, err
	}
	return u, err
}

// UpdateUser updates a user in the database
func (m *postgresDBRepo) UpdateUser(u models.User) error {
	// Create a context with a timeout of 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // Ensure the context is canceled when the function returns

	// Define an SQL query for updating a user's information
	query := `
        update users set first_name = $1, last_name = $2, email = $3, access_level = $4, updated_at = $5
    `

	// Execute the query with the provided parameters using the database connection in m.DB
	_, err := m.DB.ExecContext(ctx, query,
		u.FirstName,
		u.LastName,
		u.Email,
		u.AccessLevel,
		time.Now(), // Set the 'updated_at' field to the current time
	)

	if err != nil {
		return err // Return an error if the query execution fails
	}

	return nil // Return nil to indicate a successful update
}

// Authenticate authenticates a user
func (m *postgresDBRepo) Authenticate(email, testPassword string) (int, string, error) {
	// Create a context with a timeout of 3 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel() // Ensure the context is canceled when the function returns

	var id int
	var hashedPassword string

	// Query the database to retrieve the user's ID and hashed password by email
	row := m.DB.QueryRowContext(ctx, "select id, password from users where email = $1", email)
	err := row.Scan(&id, &hashedPassword)
	if err != nil {
		return id, "", err // If there's an error (e.g., no user found with the provided email), return an error
	}

	// Compare the provided 'testPassword' with the hashed password retrieved from the database
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(testPassword))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return 0, "", errors.New("incorrect password") // If the passwords don't match, return an error
	} else if err != nil {
		return 0, "", err // If there's any other error, return an error
	}

	return id, hashedPassword, nil // If everything is successful, return the user's ID and hashed password
}

// AllReservations returns a slice of all reservations
func (m *postgresDBRepo) AllReservations() ([]models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var reservations []models.Reservation

	query := `
		select r.id, r.first_name, r.last_name, r.email, r.phone, r.start_date, 
		r.end_date, r.room_id, r.created_at, r.updated_at, r.processed,
		rm.id, rm.room_name
		from reservations r
		left join rooms rm on (r.room_id = rm.id)
		order by r.start_date asc
`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return reservations, err
	}
	defer rows.Close()

	for rows.Next() {
		var i models.Reservation
		err := rows.Scan(
			&i.ID,
			&i.FirstName,
			&i.LastName,
			&i.Email,
			&i.Phone,
			&i.StartDate,
			&i.EndDate,
			&i.RoomID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Processed,
			&i.Room.ID,
			&i.Room.RoomName,
		)

		if err != nil {
			return reservations, err
		}
		reservations = append(reservations, i)
	}

	if err = rows.Err(); err != nil {
		return reservations, err
	}

	return reservations, nil
}

// GetReservationByID returns one reservation by ID
func (m *postgresDBRepo) GetReservationByID(id int) (models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var res models.Reservation

	query := `
		select r.id, r.first_name, r.last_name, r.email, r.phone, r.start_date,
		r.end_date, r.room_id, r.created_at, r.updated_at, r.processed, 
		rm.id, rm.room_name
		from reservations r
		left join rooms rm on (r.room_id = rm.id)
		where r.id = $1
`
	row := m.DB.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&res.ID,
		&res.FirstName,
		&res.LastName,
		&res.Email,
		&res.Phone,
		&res.StartDate,
		&res.EndDate,
		&res.RoomID,
		&res.CreatedAt,
		&res.UpdatedAt,
		&res.Processed,
		&res.Room.ID,
		&res.Room.RoomName,
	)

	if err != nil {
		return res, err
	}

	return res, nil
}

// AllNewReservations returns a slice of all reservations
func (m *postgresDBRepo) AllNewReservations() ([]models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var reservations []models.Reservation

	query := `
		select r.id, r.first_name, r.last_name, r.email, r.phone, r.start_date, 
		r.end_date, r.room_id, r.created_at, r.updated_at, 
		rm.id, rm.room_name
		from reservations r
		left join rooms rm on (r.room_id = rm.id)
		where processed = 0
		order by r.start_date asc
`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return reservations, err
	}
	defer rows.Close()

	for rows.Next() {
		var i models.Reservation
		err := rows.Scan(
			&i.ID,
			&i.FirstName,
			&i.LastName,
			&i.Email,
			&i.Phone,
			&i.StartDate,
			&i.EndDate,
			&i.RoomID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Room.ID,
			&i.Room.RoomName,
		)

		if err != nil {
			return reservations, err
		}
		reservations = append(reservations, i)
	}

	if err = rows.Err(); err != nil {
		return reservations, err
	}

	return reservations, nil
}

// UpdateReservation updates a reservation in the database
func (m *postgresDBRepo) UpdateReservation(u models.Reservation) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		update reservations set first_name = $1, last_name = $2, email = $3, phone = $4, updated_at = $5
		where id = $6
`

	_, err := m.DB.ExecContext(ctx, query,
		u.FirstName,
		u.LastName,
		u.Email,
		u.Phone,
		time.Now(),
		u.ID,
	)

	if err != nil {
		return err
	}

	return nil
}

// DeleteReservation deletes one reservation by id
func (m *postgresDBRepo) DeleteReservation(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := "delete from reservations where id = $1"

	_, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	return nil
}

// UpdateProcessedForReservation updates processed for a reservation by id
func (m *postgresDBRepo) UpdateProcessedForReservation(id, processed int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := "update reservations set processed = $1 where id = $2"

	_, err := m.DB.ExecContext(ctx, query, processed, id)
	if err != nil {
		return err
	}

	return nil
}

func (m *postgresDBRepo) InsertCustomer(res models.Customer) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	//
	var newID int
	stmt := `insert into customers (customer_code,customer_name,contact_person,contact_tel,contact_mobile,
		contact_email,customer_business,customer_location,customer_status,marketer_name,marketer_code,marketer_email,business_nature) values($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) returning customer_id`

	err := m.DB.QueryRowContext(ctx, stmt,
		res.CustomerCode,
		res.CustomerName,
		res.ContactPerson,
		res.ContactNo,
		res.MobileNo,
		res.Email,
		res.BusinessName,
		res.LocationDetails,
		res.Status,
		res.MarketerName,
		res.MarketedBy,
		res.MarketerEmail,
		res.NatureOfBusiness,
	).Scan(&newID)

	if err != nil {
		return 0, err
	}

	return newID, nil
}

// Get All Customers
func (m *postgresDBRepo) AllCustomers() ([]models.Customer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var customers []models.Customer

	query := `SELECT customer_id, customer_code, customer_name, contact_person, 
	contact_tel, contact_mobile, contact_email, customer_business, customer_location, 
	customer_status, marketer_name, marketer_code, marketer_email,business_nature
	FROM customers
`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return customers, err
	}
	defer rows.Close()

	for rows.Next() {
		var i models.Customer
		err := rows.Scan(
			&i.CustomerId,
			&i.CustomerCode,
			&i.CustomerName,
			&i.ContactPerson,
			&i.ContactNo,
			&i.MobileNo,
			&i.Email,
			&i.BusinessName,
			&i.LocationDetails,
			&i.Status,
			&i.MarketerName,
			&i.MarketedBy,
			&i.MarketerEmail,
			&i.NatureOfBusiness,
		)

		if err != nil {
			return customers, err
		}
		customers = append(customers, i)
	}

	if err = rows.Err(); err != nil {
		return customers, err
	}

	return customers, nil
}

func (m *postgresDBRepo) InsertFile(customerCode, filePath string, customerId int, uniqueFilenameWithExtension string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var newID int
	stmt := `INSERT INTO customer_images (customer_id,customer_code, file_path,file_name) VALUES ($1, $2,$3, $4) RETURNING file_id`

	err := m.DB.QueryRowContext(ctx, stmt, customerId, customerCode, filePath, uniqueFilenameWithExtension).Scan(&newID)

	if err != nil {
		return 0, err
	}

	return newID, nil
}

// GetReservationByID returns one reservation by ID
func (m *postgresDBRepo) GetCustomerByID(id int) (models.Customer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var res models.Customer

	query := `SELECT customer_id, customer_code, customer_name, contact_person, 
	contact_tel, contact_mobile, contact_email, customer_business, customer_location, 
	customer_status, marketer_name, marketer_code, marketer_email,business_nature,location_cordinates
	FROM customers where customer_id = $1`
	row := m.DB.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&res.CustomerId,
		&res.CustomerCode,
		&res.CustomerName,
		&res.ContactPerson,
		&res.ContactNo,
		&res.MobileNo,
		&res.Email,
		&res.BusinessName,
		&res.LocationDetails,
		&res.Status,
		&res.MarketerName,
		&res.MarketedBy,
		&res.MarketerEmail,
		&res.NatureOfBusiness,
		&res.LocationCoordinates,
	)

	if err != nil {
		return res, err
	}

	return res, nil
}

func (m *postgresDBRepo) GetAttachmentsByCustomerID(id int) (models.CustomerImages, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var res models.CustomerImages

	// First, retrieve the customer information
	customerQuery := `SELECT customer_id, customer_code FROM customer_images WHERE customer_id = $1`
	row := m.DB.QueryRowContext(ctx, customerQuery, id)
	err := row.Scan(&res.CustomerId, &res.CustomerCode)
	if err != nil {
		return res, err
	}

	// Now, retrieve the attachments associated with the customer
	attachmentsQuery := `SELECT file_id, file_path,file_name FROM customer_images WHERE customer_id = $1`
	rows, err := m.DB.QueryContext(ctx, attachmentsQuery, id)
	if err != nil {
		return res, err
	}
	defer rows.Close()

	// Create a slice to hold the attachments
	var attachments []models.Attachment

	for rows.Next() {
		var attachment models.Attachment
		if err := rows.Scan(&attachment.File_id, &attachment.FilePath, &attachment.FileName); err != nil {
			return res, err
		}
		attachments = append(attachments, attachment)
	}

	// Assign the attachments to the customer images
	res.Attachments = attachments

	if err := rows.Err(); err != nil {
		return res, err
	}

	return res, nil
}

// GetTradeShareInforByID returns trade license shareholder information by customer ID
// GetTradeShareInforByID returns trade license shareholder information by customer ID
func (m *postgresDBRepo) GetTradeShareInforByID(id int) ([]models.TradeLicenseHolder, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var shareholders []models.TradeLicenseHolder

	query := `SELECT trade_license_id, customer_id, shareholder_id, customer_code, shareholder_name,
	shareholder_role, shareholder_nationality, shareholder_no_of_shares, "shareholder_emirateID",
	emirateid_expire_date, shareholder_passport, passport_expire_date, id_file_path, passport_file_path, 
	created_at, updated_at
	FROM trade_license_shareholders where customer_id = $1`
	rows, err := m.DB.QueryContext(ctx, query, id)
	if err != nil {
		fmt.Println("Error executing query:", err)
		return shareholders, err
	}
	defer rows.Close()

	for rows.Next() {
		var i models.TradeLicenseHolder
		err := rows.Scan(
			&i.TradeLicenseID,
			&i.CustomerId,
			&i.ShareHolderID,
			&i.CustomerCode,
			&i.ShareHolderName,
			&i.ShareHolderRole,
			&i.ShNationality,
			&i.ShNoOfShares,
			&i.ShEmirateID,
			&i.ShEmIDExp, // Corrected the variable name here
			&i.ShPassport,
			&i.ShPassportExp, // Corrected the variable name here
			&i.ShIDFilepath,
			&i.ShPassFilepath, // Corrected the variable name here
			&i.CreatedAt,      // Added the missing fields
			&i.UpdatedAt,      // Added the missing fields
		)

		if err != nil {
			fmt.Println("Error scanning row:", err)
			return shareholders, err
		}
		shareholders = append(shareholders, i)
	}

	if err = rows.Err(); err != nil {
		fmt.Println("Error iterating through rows:", err)
		return shareholders, err
	}

	fmt.Println("Data retrieved successfully:", shareholders)
	return shareholders, nil
}

// GetReservationByID returns one reservation by ID
func (m *postgresDBRepo) GetTradeLicenseInforByID(id int) (models.TradeLicense, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var res models.TradeLicense

	query := `SELECT trade_license_id, customer_id, emirate, "mohreNo", trade_name, legal_status, establishment_date, registration_date, license_expiray, created_at, updated_at,file_path,file_name,trade_license_no
	FROM trade_license where customer_id = $1`
	row := m.DB.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&res.TradeLicenseID,
		&res.CustomerId,
		&res.Emirate,
		&res.MohreNo,
		&res.TradeName,
		&res.LegalStatus,
		&res.EstablishDate,
		&res.RegistrationDate,
		&res.LicenseExpiry,
		&res.CreatedAt,
		&res.UpdatedAt,
		&res.FilePath,
		&res.FileName,
		&res.TradeLicenseNo,
	)

	if err != nil {
		return res, err
	}

	return res, nil
}

func (m *postgresDBRepo) InsertTradeLicense(res models.TradeLicense) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	//
	var newID int
	stmt := `insert into trade_license (customer_id, emirate, 
		"mohreNo", trade_name, legal_status, establishment_date, 
		registration_date, license_expiray, created_at, updated_at, 
		file_path, file_name,trade_license_no) 
		values($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) returning trade_license_id`

	err := m.DB.QueryRowContext(ctx, stmt,
		res.CustomerId,
		res.Emirate,
		res.MohreNo,
		res.TradeName,
		res.LegalStatus,
		res.EstablishDate,
		res.RegistrationDate,
		res.LicenseExpiry,
		res.CreatedAt,
		res.UpdatedAt,
		res.FilePath,
		res.FileName,
		res.TradeLicenseNo,
	).Scan(&newID)

	if err != nil {
		return 0, err
	}

	return newID, nil
}

func (m *postgresDBRepo) GetMemorandumInforByID(id int) ([]models.Memorandum, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var memorandum []models.Memorandum

	query := `SELECT trade_license_id, customer_id, memorandum_id, customer_code, 
	representative_name, representative_no_of_shares, "representative_emirateID", 
	emirateid_expire_date, representative_passport, passport_expire_date, 
	id_file_path, passport_file_path, created_at, updated_at
	FROM memorandums where customer_id = $1`
	rows, err := m.DB.QueryContext(ctx, query, id)
	if err != nil {
		fmt.Println("Error executing query:", err)
		return memorandum, err
	}
	defer rows.Close()

	for rows.Next() {
		var i models.Memorandum
		err := rows.Scan(
			&i.TradeLicenseID,
			&i.CustomerId,
			&i.MemorandumID,
			&i.CustomerCode,
			&i.RepresentativeName,
			&i.RepNoOfShares,
			&i.RepEmID,
			&i.RepEmIDExp,
			&i.RepPassport,
			&i.RepPassportExp, // Corrected the variable name here
			&i.RepIDFilepath,
			&i.RepPassFilepath, // Corrected the variable name here
			&i.CreatedAt,       // Added the missing fields
			&i.UpdatedAt,       // Added the missing fields
		)

		if err != nil {
			fmt.Println("Error scanning row:", err)
			return memorandum, err
		}
		memorandum = append(memorandum, i)
	}

	if err = rows.Err(); err != nil {
		fmt.Println("Error iterating through rows:", err)
		return memorandum, err
	}

	fmt.Println("Data retrieved successfully:", memorandum)
	return memorandum, nil
}

func (m *postgresDBRepo) InsertPartner(res models.TradeLicenseHolder) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var newID int
	stmt := `insert into trade_license_shareholders (trade_license_id, customer_id, 
		customer_code,
			shareholder_name, shareholder_role, 
			shareholder_nationality, shareholder_no_of_shares, 
			"shareholder_emirateID", emirateid_expire_date, 
			shareholder_passport, passport_expire_date, 
			id_file_path, passport_file_path,created_at,updated_at) values($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15) returning shareholder_id`

	err := m.DB.QueryRowContext(ctx, stmt,
		res.TradeLicenseID,
		res.CustomerId,
		res.CustomerCode,
		res.ShareHolderName,
		res.ShareHolderRole,
		res.ShNationality,
		res.ShNoOfShares,
		res.ShEmirateID,
		res.ShEmIDExp,
		res.ShPassport,
		res.ShPassportExp,
		res.ShIDFilepath,
		res.ShPassFilepath,
		res.CreatedAt,
		res.UpdatedAt,
	).Scan(&newID)

	if err != nil {
		fmt.Println("Error inserting partner:", err)
		return 0, err
	}

	fmt.Println("Inserted new partner with ID:", newID)
	return newID, nil
}

func (m *postgresDBRepo) GetCustomerCodeByID(id int) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT customer_code FROM customers WHERE customer_id = $1`

	var customerCode string
	err := m.DB.QueryRowContext(ctx, query, id).Scan(&customerCode)
	if err != nil {
		return "", err
	}

	return customerCode, err
}

func (m *postgresDBRepo) InsertMemorandum(res models.Memorandum) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var newID int
	stmt := `insert into memorandums 
	(trade_license_id, customer_id, 
		customer_code, representative_name, representative_no_of_shares, 
		"representative_emirateID", emirateid_expire_date, 
		representative_passport, passport_expire_date, id_file_path,
		passport_file_path, created_at, updated_at) 
		values($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) returning memorandum_id`

	err := m.DB.QueryRowContext(ctx, stmt,
		res.TradeLicenseID,
		res.CustomerId,
		res.CustomerCode,
		res.RepresentativeName,
		res.RepNoOfShares,
		res.RepEmID,
		res.RepEmIDExp,
		res.RepPassport,
		res.RepPassportExp,
		res.RepIDFilepath,
		res.RepPassFilepath,
		res.CreatedAt,
		res.UpdatedAt,
	).Scan(&newID)

	if err != nil {
		fmt.Println("Error inserting partner:", err)
		return 0, err
	}

	fmt.Println("Inserted new partner with ID:", newID)
	return newID, nil
}
