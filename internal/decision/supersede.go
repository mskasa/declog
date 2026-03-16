package decision

import (
	"fmt"
	"strings"
)

// CheckSupersedable returns an error if the decision cannot be superseded
// because it is already inactive or superseded.
func CheckSupersedable(d *Decision) error {
	status := d.Status
	if strings.EqualFold(status, "Inactive") {
		return fmt.Errorf("ADR %04d is already Inactive", d.ID)
	}
	if strings.HasPrefix(strings.ToLower(status), "superseded") {
		return fmt.Errorf("ADR %04d is already %s", d.ID, status)
	}
	return nil
}
