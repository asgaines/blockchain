package mining

import "errors"

type HashSpeed int

func ToSpeed(s string) (HashSpeed, error) {
	switch s {
	case "low":
		return LowSpeed, nil
	case "medium":
		return MediumSpeed, nil
	case "high":
		return HighSpeed, nil
	case "ultra":
		return UltraSpeed, nil
	}

	return 0, errors.New("invalid speed")
}

const (
	LowSpeed HashSpeed = iota + 1
	MediumSpeed
	HighSpeed
	UltraSpeed
)
