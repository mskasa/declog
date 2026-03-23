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
		return fmt.Errorf("document %q is already Inactive", d.Slug)
	}
	if strings.HasPrefix(strings.ToLower(status), "superseded") {
		return fmt.Errorf("document %q is already %s", d.Slug, status)
	}
	return nil
}
