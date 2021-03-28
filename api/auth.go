package api

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	guuid "github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Initalize and set the authentication and authorization routes
func AuthRoutes(router fiber.Router, db *gorm.DB) {
	auth := router.Group("/auth", SecurityMiddleware)
	auth.Post("/login", func(c *fiber.Ctx) error {
		json := new(User)
		if err := c.BodyParser(json); err != nil {
			return c.SendStatus(500)
		}
		empty := User{}
		if json.Username == empty.Username || empty.Password == json.Password {
			return c.Status(401).SendString("Invalid Data Sent")
		}

		foundUser := User{}
		queryUser := User{Username: json.Username}
		err := db.First(&foundUser, &queryUser).Error
		if err == gorm.ErrRecordNotFound {
			return c.Status(401).SendString("User not Found")
		}
		if !comparePasswords(foundUser.Password, []byte(json.Password)) {
			return c.Status(401).SendString("Incorrect Password")
		}
		newSession := Session{UserRefer: foundUser.ID, Sessionid: guuid.New()}
		CreateErr := db.Create(&newSession).Error
		if CreateErr != nil {
			return c.Status(500).SendString("Creation Error")
		}
		c.Cookie(&fiber.Cookie{
			Name:     "sessionid",
			Expires:  time.Now().Add(5 * 24 * time.Hour),
			Value:    newSession.Sessionid.String(),
			HTTPOnly: true,
		})
		return c.Status(200).JSON(newSession)
	})

	auth.Post("/logout", func(c *fiber.Ctx) error {
		json := new(Session)
		if err := c.BodyParser(json); err != nil {
			return c.SendStatus(500)
		}
		if json.Sessionid == new(Session).Sessionid {
			return c.Status(401).SendString("Invalid Data Sent")
		}
		session := Session{}
		query := Session{Sessionid: json.Sessionid}
		err := db.First(&session, &query).Error
		if err == gorm.ErrRecordNotFound {
			return c.Status(401).SendString("Session Not Found")
		}
		db.Delete(&session)
		c.ClearCookie("sessionid")
		return c.SendStatus(200)
	})
	auth.Post("/create", func(c *fiber.Ctx) error {
		json := new(User)
		if err := c.BodyParser(json); err != nil {
			return c.SendStatus(500)
		}
		empty := User{}
		if json.Username == empty.Username || empty.Password == json.Password {
			return c.Status(401).SendString("Invalid Data Sent")
		}
		pw := hashAndSalt([]byte(json.Password))
		newUser := User{
			Username: json.Username,
			Password: pw,
			ID:       guuid.New(),
		}
		foundUser := User{}
		query := User{Username: json.Username}
		err := db.First(&foundUser, &query).Error
		if err != gorm.ErrRecordNotFound {
			return c.Status(401).SendString("User Already Exists")
		}
		db.Create(&newUser)
		return c.SendStatus(200)
	})
	auth.Post("/user", func(c *fiber.Ctx) error {
		user := User{}
		myUser := User{Username: "NikSchaefer"}
		Sessions := []Session{}
		db.First(&user, &myUser)
		db.Model(&user).Association("Sessions").Find(&Sessions)
		user.Sessions = Sessions
		return c.JSON(user)
	})
	auth.Post("/delete", func(c *fiber.Ctx) error {
		json := new(User)
		if err := c.BodyParser(json); err != nil {
			return c.SendStatus(400)
		}
		empty := User{}
		if json.Username == empty.Username || empty.Password == json.Password {
			return c.Status(401).SendString("Invalid Data Sent")
		}
		foundUser := User{}
		query := User{Username: json.Username}
		err := db.First(&foundUser, &query).Error
		if err == gorm.ErrRecordNotFound {
			return c.Status(401).SendString("User Not Found")
		}
		if !comparePasswords(foundUser.Password, []byte(json.Password)) {
			return c.Status(401).SendString("Invalid Credentials")
		}
		db.Model(&foundUser).Association("Sessions").Clear()
		createErr := db.Delete(&foundUser).Error
		if createErr != nil {
			fmt.Println(createErr)
		}
		c.ClearCookie("sessionid")
		return c.SendStatus(200)
	})
	auth.Post("/update", func(c *fiber.Ctx) error {
		json := new(User)
		if err := c.BodyParser(json); err != nil {
			return c.SendStatus(500)
		}
		empty := User{}
		if json.Username == empty.Username || empty.Password == json.Password {
			return c.Status(401).SendString("Invalid Data Sent")
		}
		foundUser := User{}
		query := User{Username: json.Username}
		err := db.First(&foundUser, &query).Error
		if err == gorm.ErrRecordNotFound {
			return c.Status(401).SendString("User Not Found")
		}
		return c.SendStatus(200)
	})
	auth.Post("/changepassword", func(c *fiber.Ctx) error {
		json := new(ChangePassword)
		if err := c.BodyParser(json); err != nil {
			return c.SendStatus(400)
		}
		foundUser := User{}
		query := User{Username: json.Username}
		err := db.First(&foundUser, &query).Error
		if err == gorm.ErrRecordNotFound {
			return c.Status(401).SendString("User Not Found")
		}
		if !comparePasswords(foundUser.Password, []byte(json.NewPassword)) {
			return c.Status(401).SendString("Invalid Password")
		}
		foundUser.Password = hashAndSalt([]byte(json.Password))
		db.Save(&foundUser)
		return c.SendStatus(200)
	})
}

func hashAndSalt(pwd []byte) string {
	// Use GenerateFromPassword to hash & salt pwd.
	// MinCost is just an integer constant provided by the bcrypt
	// package along with DefaultCost & MaxCost.
	// The cost can be any value you want provided it isn't lower
	// than the MinCost (4)
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	if err != nil {
		fmt.Println(err)
	}
	// GenerateFromPassword returns a byte slice so we need to
	// convert the bytes to a string and return it
	return string(hash)
}
func comparePasswords(hashedPwd string, plainPwd []byte) bool {
	// Since we'll be getting the hashed password from the DB it
	// will be a string so we'll need to convert it to a byte slice
	byteHash := []byte(hashedPwd)
	err := bcrypt.CompareHashAndPassword(byteHash, plainPwd)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}