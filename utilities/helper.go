package utilities

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"image/jpeg"
	"image/png"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

// BooleanPointer :
func BooleanPointer(i bool) *bool {
	return &i
}

// IntPointer :
func IntPointer(i int) *int {
	return &i
}

// FloatPointer :
func FloatPointer(i float64) *float64 {
	return &i
}

// StringPointer :
func StringPointer(i string) *string {
	return &i
}

// RandomString
func RandomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length)
	rand.Read(b)
	return fmt.Sprintf("%x", b)[:length]
}

func RandomInt(min, max int) int {
	if min >= max {
		return min
	}
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min+1) + min
}

// PhoneNumberTrim
func PhoneNumberTrim(phoneNumber string) (*string, error) {
	phoneNumber = strings.ReplaceAll(phoneNumber, "+62", "")
	phoneNumber = strings.ReplaceAll(phoneNumber, "+", "")
	phoneNumber = strings.ReplaceAll(phoneNumber, "-", "")
	phoneNumber = strings.ReplaceAll(phoneNumber, ".", "")
	phoneNumber = strings.ReplaceAll(phoneNumber, "#", "")
	phoneNumber = strings.ReplaceAll(phoneNumber, "*", "")
	phoneNumber = strings.ReplaceAll(phoneNumber, ".", "")
	runePhoneNumber := []rune(phoneNumber)
	getFirstCharacter := string(runePhoneNumber[0:1])
	if getFirstCharacter == "0" {
		_, i := utf8.DecodeRuneInString(phoneNumber)
		phoneNumber = phoneNumber[i:]
	}

	if !strings.HasPrefix(phoneNumber, "8") {
		return nil, fmt.Errorf(ValueNotValid, "Nomor Telepon")

	}

	return &phoneNumber, nil
}

// GetLastDateOfMonth
func GetLastDateOfMonth(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month()+1, 1, 0, 0, 0, 0, date.Location()).AddDate(0, 0, -1)
}

// GetFirstDateOfMonth
func GetFirstDateOfMonth(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
}

// NumberFormat
func NumberFormat(number float64) string {
	str := fmt.Sprintf("%v", number)
	re := regexp.MustCompile("(\\d+)(\\d{3})")
	for n := ""; n != str; {
		n = str
		str = re.ReplaceAllString(str, "$1.$2")
	}
	return str
}

func Slug(slug string) string {
	slug = strings.ToLower(slug)
	slug = strings.ReplaceAll(slug, " ", "-")
	return slug
}

func Upload(file string) (*string, error) {

	idx := strings.Index(file, ";base64,")
	if idx < 0 {
		return nil, fmt.Errorf(ValueNotValid, "image")
	}

	unbased, err := base64.StdEncoding.DecodeString(file[idx+8:])
	if err != nil {
		return nil, fmt.Errorf(ValueNotValid, "image")
	}

	var (
		imageType = file[11:idx]
		fileName  = fmt.Sprintf("%s.%s", RandomString(15), imageType)
	)

	if imageType != "png" && imageType != "jpeg" && imageType != "jpg" {
		return nil, fmt.Errorf(ValueNotValid, "image")
	} else {
		fileOpen, err := os.OpenFile(fmt.Sprintf("uploads/%s", fileName), os.O_WRONLY|os.O_CREATE, 0777)
		if err != nil {
			return nil, err
		}

		switch imageType {
		case "png":
			reader := bytes.NewReader(unbased)
			img, err := png.Decode(reader)
			if err != nil {
				return nil, fmt.Errorf(ValueNotValid, "image")
			}

			png.Encode(fileOpen, img)
		case "jpeg", "jpg":
			reader := bytes.NewReader(unbased)
			img, err := jpeg.Decode(reader)
			if err != nil {
				return nil, fmt.Errorf(ValueNotValid, "image")
			}

			jpeg.Encode(fileOpen, img, nil)
		default:
			return nil, errors.New(InternalServiceError)
		}
	}

	return &fileName, nil
}

func DereferenceString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func GeneratePrefixedUUID(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, uuid.NewString())
}

func CalculateDisplayPrice(price float64, taxPercentage *float64, taxInclusive *bool) float64 {
	if taxPercentage == nil || taxInclusive == nil {
		return price
	}

	if *taxInclusive {
		return price
	}

	return price + (price * (*taxPercentage / 100))
}
