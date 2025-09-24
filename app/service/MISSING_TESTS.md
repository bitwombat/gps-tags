# Missing Test Cases for GPS Tags Service

This document outlines missing test cases that should be implemented to improve test coverage and ensure robustness of the GPS tags service.

## Current Test Coverage Analysis

**Existing Tests:**
- `handledata_test.go`: Basic data upload happy path with battery/zone notifications, authentication scenarios
- `handlepages_test.go`: Map page generation (current and paths)
- Component tests exist for: oneshot, zones, storage, device packages

**Missing Coverage:** The following areas lack comprehensive testing and need additional test cases.

---

## 1. Data Upload Handler (`handledata.go`) Missing Tests

### 1.1 Battery Notification Edge Cases

#### Issue: Test Battery Voltage Boundary Conditions
**Description:** Test battery notification behavior at exact threshold values and hysteresis boundaries.
**Test Cases Needed:**
- Battery voltage exactly at low threshold (4.0V) - should trigger notification
- Battery voltage exactly at critical threshold (3.8V) - should trigger critical notification
- Battery voltage at hysteresis boundary (4.1V) - should reset low battery notification
- Battery voltage between critical and low thresholds (3.9V) - should trigger critical but not low
- Battery voltage above low threshold but below reset (4.05V) - should not reset notifications

**Implementation Notes:**
- Create test data with `AnalogueData.1` values: 4000, 3800, 4100, 3900, 4050
- Use midday time (hour 13) to ensure notifications trigger
- Verify correct notification titles and messages

#### Issue: Test Battery Notifications During Night Hours
**Description:** Battery notifications should be suppressed between 22:00 and 08:00.
**Test Cases Needed:**
- Low battery at 23:00 (should not notify)
- Low battery at 02:00 (should not notify)
- Low battery at 07:59 (should not notify)
- Low battery at 08:00 (should notify)
- Low battery at 22:00 (should notify)
- Low battery at 22:01 (should not notify)

**Implementation Notes:**
- Test with battery voltage 3.5V (below both thresholds)
- Verify FakeNotifier.notifications slice is empty for night hours
- Verify notifications are sent during waking hours

#### Issue: Test Missing Analogue Reading
**Description:** Handler should gracefully handle transmissions without analogue data.
**Test Cases Needed:**
- Transmission with GPS data but no AnalogueReading
- Transmission with multiple records, some with and some without AnalogueReading

**Implementation Notes:**
- Remove AnalogueReading field from test JSON
- Verify no battery notifications are sent
- Verify no errors or panics occur

### 1.2 Zone Notification Scenarios

#### Issue: Test All Zone Boundary Scenarios
**Description:** Test dog location relative to property and safe zone boundaries.
**Test Cases Needed:**
- Dog inside both property and safe zone (no notifications)
- Dog inside property but outside safe zone (safe zone notification only)
- Dog outside property but inside safe zone (property notification only)
- Dog outside both boundaries (both notifications)
- Dog returning to property (reset notification)
- Dog returning to safe zone (reset notification)

**Implementation Notes:**
- Use coordinates from `geodata.go` property and safe zone boundaries
- Test coordinates: inside property (-31.4580, 152.6420), outside property (-31.4500, 152.6500)
- Verify correct notification titles and zone text messages

#### Issue: Test Missing GPS Reading
**Description:** Handler should gracefully handle transmissions without GPS data.
**Test Cases Needed:**
- Transmission with analogue data but no GPSReading
- Transmission with multiple records, some with and some without GPSReading

**Implementation Notes:**
- Remove GPSReading field from test JSON
- Verify no zone notifications are sent
- Verify no errors or panics occur

#### Issue: Test Named Zones Nil Scenario
**Description:** Test behavior when named zones KML files cannot be loaded.
**Test Cases Needed:**
- Handler created with failed zone loading (simulate KML read error)
- Verify zone text shows "<No zones loaded>"

**Implementation Notes:**
- Mock zonespkg.ReadKMLDir to return error
- Verify notification message contains "<No zones loaded>"

### 1.3 GPS Data Validation

#### Issue: Test Invalid GPS Coordinates
**Description:** Handler should reject GPS coordinates of 0,0 as invalid.
**Test Cases Needed:**
- GPS data with Lat=0, Long=152.6420 (should be rejected)
- GPS data with Lat=-31.4580, Long=0 (should be rejected)
- GPS data with Lat=0, Long=0 (should be rejected)
- Valid GPS data mixed with invalid in same transmission

**Implementation Notes:**
- Verify records with 0,0 coordinates are not processed
- Verify error log message "Got 0 for lat or long... not committing record"
- Verify valid records in same transmission are still processed

### 1.4 HTTP Method and Error Handling

#### Issue: Test Non-POST HTTP Methods
**Description:** Upload endpoint should only accept POST requests.
**Test Cases Needed:**
- GET request to /upload (should return 405 Method Not Allowed)
- PUT request to /upload (should return 405 Method Not Allowed)
- DELETE request to /upload (should return 405 Method Not Allowed)

**Implementation Notes:**
- Verify HTTP status 405 is returned
- Verify error is logged "Got a request to /upload that was not a POST"

#### Issue: Test Storage Write Failures
**Description:** Test behavior when database write operations fail.
**Test Cases Needed:**
- Storage WriteTx returns error
- Verify HTTP 500 status returned
- Verify error logged

**Implementation Notes:**
- Mock FakeStorer.WriteTx to return error
- Verify HTTP status 500 and error logging

#### Issue: Test Malformed Request Data
**Description:** Test behavior with invalid JSON and request body issues.
**Test Cases Needed:**
- Invalid JSON in request body
- Empty request body
- Request body read error (simulated)

**Implementation Notes:**
- Send malformed JSON string
- Verify HTTP 500 status and error logging
- Mock io.ReadAll to return error

#### Issue: Test Context Timeout
**Description:** Test behavior when request context times out.
**Test Cases Needed:**
- Request that exceeds 20 second timeout
- Storage operation that blocks beyond timeout

**Implementation Notes:**
- Use context with short timeout in test
- Verify proper cleanup and error handling

### 1.5 Serial Number Handling

#### Issue: Test Unknown Tag Serial Numbers
**Description:** Handler should log warnings for unknown tag serial numbers.
**Test Cases Needed:**
- Data from tag with unknown SerNo (not in model.SerNoToName map)
- Verify warning is logged "Unknown tag number: X"
- Verify processing continues with empty dog name

**Implementation Notes:**
- Use SerNo not in map (e.g., 999999)
- Verify warning log output
- Verify notifications still work (with empty/fallback name)

---

## 2. Page Handlers (`handlepages.go`) Missing Tests

### 2.1 Error Handling Scenarios

#### Issue: Test Storage Read Failures
**Description:** Test behavior when database read operations fail.
**Test Cases Needed:**
- GetLastStatuses returns error (current map page)
- GetLastNCoords returns error (paths map page)
- Verify HTTP 500 status returned for both cases

**Implementation Notes:**
- Mock storage methods to return errors
- Verify HTTP status 500 and error logging

#### Issue: Test Template File Errors
**Description:** Test behavior when HTML template files are missing or corrupted.
**Test Cases Needed:**
- Missing current-map.html file
- Missing paths.html file
- Corrupted template files (invalid substitution markers)

**Implementation Notes:**
- Test in environment where template files don't exist
- Verify HTTP 500 status and error logging

#### Issue: Test Response Write Failures
**Description:** Test behavior when writing HTTP response fails.
**Test Cases Needed:**
- Response writer that fails on Write() call
- Verify error logging occurs

**Implementation Notes:**
- Create mock ResponseWriter that returns error on Write()
- Verify error is logged "Error writing response"

### 2.2 Data Edge Cases

#### Issue: Test Empty/Minimal Data Sets
**Description:** Test page generation with minimal or empty data.
**Test Cases Needed:**
- Empty Statuses map (no tags)
- Empty Coords map (no paths)
- Single tag with minimal data
- Very large coordinate datasets

**Implementation Notes:**
- Test with empty storage responses
- Verify pages render without errors
- Test performance with large datasets

---

## 3. Health and Notification Handlers (`handleothers.go`) Missing Tests

### 3.1 Health Check Endpoint

#### Issue: Test Health Check Endpoint
**Description:** Complete testing of /health endpoint behavior.
**Test Cases Needed:**
- GET request returns 200 OK with "OK" body
- Health check logging behavior (first call logs, subsequent calls don't)
- Multiple consecutive health checks
- POST/PUT requests to health check (should still work)

**Implementation Notes:**
- Verify response status 200 and body "OK"
- Verify logging behavior with lastWasHealthCheck flag
- Test all HTTP methods

### 3.2 Test Notification Endpoints

#### Issue: Test Notification Endpoints
**Description:** Test all notification testing endpoints (/testnotify, /notifytest, etc.).
**Test Cases Needed:**
- Successful notification send
- Notifier.Notify returns error (should return HTTP 500)
- Context timeout during notification
- Test all endpoint aliases (/testnotify, /notifytest, /testnotification, /notificationtest)

**Implementation Notes:**
- Mock notifier to return error
- Verify HTTP 500 status on notifier failure
- Test each endpoint alias behaves identically

---

## 4. Main Application (`main.go`) Missing Tests

### 4.1 Environment Variable Validation

#### Issue: Test Environment Variable Handling
**Description:** Test application startup with various environment variable configurations.
**Test Cases Needed:**
- Missing TAG_AUTH_KEY (should exit with code 1)
- Missing NTFY_SUBSCRIPTION_ID (should log warning, continue)
- NONOTIFY environment variable set (should use null notifier)
- All environment variables properly set

**Implementation Notes:**
- Test run() function return values
- Verify warning logs for missing NTFY_SUBSCRIPTION_ID
- Verify null notifier usage with NONOTIFY

### 4.2 Hostname-Based File Serving

#### Issue: Test File Server Hostname Routing
**Description:** Test hostnameBasedFileServer function with different hostnames.
**Test Cases Needed:**
- Request to tags.bitwombat.com.au (should serve from ./public_html)
- Request to photos.bitwombat.com.au (should serve from ./public_html.photos)
- Request to unknown.example.com (should serve from ./public_html default)
- Case-insensitive hostname handling
- Hostname with port numbers

**Implementation Notes:**
- Create test HTTP requests with different Host headers
- Verify correct directory selection
- Test case sensitivity (hosts should be lowercased)

### 4.3 Server Startup Scenarios

#### Issue: Test HTTP Server Configuration
**Description:** Test server startup and configuration scenarios.
**Test Cases Needed:**
- Server binding to correct ports (80, 443)
- ReadHeaderTimeout configuration (10 seconds)
- Error group handling for multiple servers
- Server startup failure scenarios

**Implementation Notes:**
- Test server configuration without actual network binding
- Mock server startup failures
- Verify errgroup usage

---

## 5. Utility Functions (`utils.go`) Missing Tests

### 5.1 Error Logging Utility

#### Issue: Test logIfErr Function
**Description:** Test error logging utility function.
**Test Cases Needed:**
- logIfErr with nil error (should not log)
- logIfErr with actual error (should log to errorLogger)
- Verify correct error message format

**Implementation Notes:**
- Test with mock logger to capture output
- Verify nil errors don't produce log output
- Verify error messages are properly formatted

---

## 6. Integration Tests Missing

### 6.1 End-to-End Workflow Tests

#### Issue: Full Request Processing Integration Tests
**Description:** Test complete request processing workflows.
**Test Cases Needed:**
- Complete data upload workflow with notifications
- Map page generation after data upload
- Multiple concurrent uploads
- Server lifecycle (startup, request processing, shutdown)

**Implementation Notes:**
- Use real HTTP server in tests
- Test concurrent request handling
- Verify data persistence across requests

### 6.2 Database Integration Tests

#### Issue: Real Database Integration Tests
**Description:** Test with real SQLite database instead of mocks.
**Test Cases Needed:**
- Data persistence across application restarts
- Database migration scenarios
- Concurrent database access
- Database corruption recovery

**Implementation Notes:**
- Use temporary database files
- Test actual SQL operations
- Verify data integrity

---

## 7. Performance and Load Tests Missing

### 7.1 Performance Benchmarks

#### Issue: Performance Benchmark Tests
**Description:** Add benchmark tests for critical paths.
**Test Cases Needed:**
- Data upload processing performance
- Map page generation performance
- Concurrent request handling
- Memory usage patterns

**Implementation Notes:**
- Use Go benchmark testing
- Test with varying data sizes
- Monitor memory allocations

---

## Implementation Priority

**High Priority (Critical for robustness):**
1. Battery notification edge cases (night hours, thresholds)
2. Zone notification scenarios
3. HTTP error handling (method validation, storage failures)
4. Environment variable validation
5. Health check and notification endpoints

**Medium Priority (Important for completeness):**
1. GPS data validation edge cases
2. Template and file serving errors
3. Hostname-based routing
4. Error logging utilities

**Lower Priority (Nice to have):**
1. Integration tests
2. Performance benchmarks
3. Database integration tests

Each test should follow the existing pattern using testify assertions and httptest for HTTP handlers.