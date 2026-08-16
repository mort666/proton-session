/**
 * Copyright © 2020-2026 Stephen Kapp and Reaper Technologies Limited.
 * All Rights Reserved.
 *
 * @Author: Stephen Kapp
 * @Date: 2026-8-16 00:58:57
 * @Last Modified by: Stephen Kapp
 * @Last Modified time: 2026-8-16 00:58:57
 */
package webclient

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"rtlabs.tech/protonsession/pkg/errors"
	"rtlabs.tech/protonsession/pkg/proton"
)

// PromptHvURL prints the human verification URL for the user to complete in a
// browser and then blocks, via [WebApiClient.WaitForEnter], until the user
// presses ENTER on the console to confirm the challenge was completed.
//
// It returns [errors.ErrHVInputTimeoutError] when no input was received within
// the configured timeout (see [WebApiClient.HVInputTimeout]), or the context
// error if ctx was cancelled first.
func (c *WebApiClient) PromptHvURL(ctx context.Context, details *proton.APIHVDetails) error {
	hvURL := c.FormatHvURL(details)

	logger := log.With().Str("hvURL", hvURL).Strs("hvMethods", details.Methods).Str("hvToken",details.Token).Logger()

	logger.Info().Msg("HV request details")

	fmt.Print("\nHuman Verification requested. Please open the URL below in a browser and press ENTER when the challenge has been completed.\n\n", hvURL+"\n")

	if err := c.WaitForEnter(ctx); err != nil {
		if errors.Is(err, errors.ErrHVInputTimeoutError) {
			fmt.Println("\nTimed out waiting for human verification confirmation.")
		}
		return err
	}

	fmt.Println("Authenticating ...")

	return nil
}

// WaitForEnter implements a Readline like wait for the user to press ENTER on
// the console, if console input is available at all.
//
// The wait is bounded by [WebApiClient.HVInputTimeout] (falling back to
// [DefaultHVInputTimeout] when unset) and by ctx. If the timeout elapses, or
// stdin is unavailable / closed before any line was read, it returns
// [errors.ErrHVInputTimeoutError] to indicate to the caller that no input was
// received. If ctx is cancelled first, the context error is returned.
//
// Any text typed before ENTER is discarded; only the ENTER keypress matters.
func (c *WebApiClient) WaitForEnter(ctx context.Context) error {
	timeout := c.HVInputTimeout
	if timeout <= 0 {
		timeout = DefaultHVInputTimeout
	}

	logger := c.hvLogger()

	// Bail out early when there is nothing sensible to read from, rather than
	// blocking a caller (e.g. a daemon or CI run) forever on a dead stdin.
	if !stdinIsReadable() {
		logger.Warn().Msg("console input unavailable while waiting for ENTER")
		return errors.ErrHVInputTimeoutError
	}

	logger.Info().Dur("timeout", timeout).Msg("waiting for ENTER on console")

	deadline := time.Now().Add(timeout)

	// Preferred path: let the runtime poller enforce the deadline so no reader
	// goroutine is left behind when the user never presses ENTER.
	if err := os.Stdin.SetReadDeadline(deadline); err == nil {
		defer os.Stdin.SetReadDeadline(time.Time{}) //nolint:errcheck

		done := make(chan error, 1)
		go func() {
			_, readErr := bufio.NewReader(os.Stdin).ReadString('\n')
			done <- readErr
		}()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case readErr := <-done:
			return mapReadResult(logger, readErr)
		}
	}

	// Fallback path: stdin does not support deadlines (non pollable file
	// descriptor), so race a blocking read against a timer. The read goroutine
	// lingers until input eventually arrives, which is an accepted trade-off
	// for a blocking os.Stdin read in Go.
	done := make(chan error, 1)
	go func() {
		_, readErr := bufio.NewReader(os.Stdin).ReadString('\n')
		done <- readErr
	}()

	timer := time.NewTimer(time.Until(deadline))
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		logger.Warn().Dur("timeout", timeout).Msg("timed out waiting for ENTER on console")
		return errors.ErrHVInputTimeoutError
	case readErr := <-done:
		return mapReadResult(logger, readErr)
	}
}

// mapReadResult translates the result of the stdin read into the error
// contract of [WebApiClient.WaitForEnter].
func mapReadResult(logger zerolog.Logger, readErr error) error {
	switch {
	case readErr == nil:
		return nil
	case errors.Is(readErr, os.ErrDeadlineExceeded):
		logger.Warn().Msg("timed out waiting for ENTER on console")
		return errors.ErrHVInputTimeoutError
	case errors.Is(readErr, io.EOF):
		// stdin closed without a newline: treat as "no input received".
		logger.Warn().Msg("console input closed while waiting for ENTER")
		return errors.ErrHVInputTimeoutError
	default:
		return fmt.Errorf("reading console input: %w", readErr)
	}
}

// stdinIsReadable reports whether stdin looks like something we can actually
// read a line from (a terminal, a pipe, or a non empty redirected file).
func stdinIsReadable() bool {
	info, err := os.Stdin.Stat()
	if err != nil {
		return false
	}

	mode := info.Mode()
	if mode&os.ModeCharDevice != 0 || mode&os.ModeNamedPipe != 0 || mode&os.ModeSocket != 0 {
		return true
	}

	return info.Size() > 0
}

// hvLogger returns the client logger, or a no-op logger when none was
// configured, so console waiting never panics on a nil logger.
func (c *WebApiClient) hvLogger() zerolog.Logger {
	if c.Logger != nil {
		return *c.Logger
	}
	return zerolog.Nop()
}


const (
	ExtractionErrorMsg         = "Human verification requested, but an issue occurred. Please try again."
	VerificationFailedErrorMsg = "Human verification failed. Please try again."
)

// VerifyAndExtractHvRequest expects an error request as input
// determines whether the given error is a Proton human verification request; if it isn't then it returns -> nil, nil (no details, no error)
// if it is a HV req. then it tries to parse the json data and verify that the captcha method is included; if either fails -> nil, err
// if the HV request was successfully decoded and the preconditions were met it returns the hv details -> hvDetails, nil.
func (c *WebApiClient) VerifyAndExtractHvRequest(err error) (*proton.APIHVDetails, error) {
	var hvDetails *proton.APIHVDetails
	var hverr error

	if err == nil {
		return nil, nil
	}

	protonErr, ok := errors.AsType[*proton.APIError](err)
	if !ok {
		return nil, err
	}

	if protonErr.IsHVError() {
		hvDetails, hverr = protonErr.GetHVDetails()
		if hverr != nil {
			return nil, hverr
		}
	}

	return hvDetails, nil
}

func (c* WebApiClient) IsHvRequest(err error) bool {
	if err == nil {
		return false
	}

	protonErr, ok := errors.AsType[*errors.APIError](err) 
	if ok && protonErr.IsHVError() {
		return true
	}

	return false
}

func (c *WebApiClient)  FormatHvURL(details *proton.APIHVDetails) string {
	return fmt.Sprintf("https://verify.proton.me/?methods=%v&token=%v",
		strings.Join(details.Methods, ","),
		details.Token)
}

const HvPMTokenHeaderField = "x-pm-human-verification-token"
const HvPMTokenType = "x-pm-human-verification-token-type"

func (c *WebApiClient) AddHVToRequest(req *http.Request, hv *proton.APIHVDetails) *http.Request {
	if hv == nil {
		return req
	}

	req.Header.Set(HvPMTokenHeaderField, hv.Token)
	req.Header.Set(HvPMTokenType, strings.Join(hv.Methods, ","))

	return req
}


