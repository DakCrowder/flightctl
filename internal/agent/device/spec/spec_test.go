package spec

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/flightctl/flightctl/api/v1alpha1"
	"github.com/flightctl/flightctl/internal/agent/client"
	"github.com/flightctl/flightctl/internal/agent/device/fileio"
	"github.com/flightctl/flightctl/internal/container"
	"github.com/flightctl/flightctl/pkg/log"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"k8s.io/apimachinery/pkg/util/wait"
)

func TestBootstrapCheckRollback(t *testing.T) {
	require := require.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReadWriter := fileio.NewMockReadWriter(ctrl)
	mockBootcClient := container.NewMockBootcClient(ctrl)

	s := &SpecManager{
		log:              log.NewPrefixLogger("test"),
		deviceReadWriter: mockReadWriter,
		bootcClient:      mockBootcClient,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Run("no rollback: bootstrap case empty desired spec", func(t *testing.T) {
		wantIsRollback := false
		mockReadWriter.EXPECT().ReadFile(gomock.Any()).Return([]byte(`{}`), nil)

		isRollback, err := s.IsRollingBack(ctx)
		require.NoError(err)
		require.Equal(wantIsRollback, isRollback)
	})

	t.Run("no rollback: booted os is equal to desired", func(t *testing.T) {
		wantIsRollback := false
		rollbackImage := "flightctl-device:v1"
		bootedImage := "flightctl-device:v2"
		desiredImage := "flightctl-device:v2"

		// desiredSpec
		desiredSpec, err := createTestSpec(desiredImage)
		require.NoError(err)
		mockReadWriter.EXPECT().ReadFile(gomock.Any()).Return(desiredSpec, nil)

		// rollbackSpec
		rollbackSpec, err := createTestSpec(rollbackImage)
		require.NoError(err)
		mockReadWriter.EXPECT().ReadFile(gomock.Any()).Return(rollbackSpec, nil)

		// bootcStatus
		bootcStatus := &container.BootcHost{}
		bootcStatus.Status.Booted.Image.Image.Image = bootedImage
		mockBootcClient.EXPECT().Status(ctx).Return(bootcStatus, nil)

		isRollback, err := s.IsRollingBack(ctx)
		require.NoError(err)
		require.Equal(wantIsRollback, isRollback)
	})

	t.Run("rollback case: rollback os equal to booted os but not desired", func(t *testing.T) {
		wantIsRollback := true
		rollbackImage := "flightctl-device:v1"
		bootedImage := "flightctl-device:v1"
		desiredImage := "flightctl-device:v2"

		// desiredSpec
		desiredSpec, err := createTestSpec(desiredImage)
		require.NoError(err)
		mockReadWriter.EXPECT().ReadFile(gomock.Any()).Return(desiredSpec, nil)

		// rollbackSpec
		rollbackSpec, err := createTestSpec(rollbackImage)
		require.NoError(err)
		mockReadWriter.EXPECT().ReadFile(gomock.Any()).Return(rollbackSpec, nil)

		// bootcStatus
		bootcStatus := &container.BootcHost{}
		bootcStatus.Status.Booted.Image.Image.Image = bootedImage
		mockBootcClient.EXPECT().Status(ctx).Return(bootcStatus, nil)

		isRollback, err := s.IsRollingBack(ctx)
		require.NoError(err)
		require.Equal(wantIsRollback, isRollback)
	})

}

func TestInitialize(t *testing.T) {
	require := require.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReadWriter := fileio.NewMockReadWriter(ctrl)

	s := &SpecManager{
		log:              log.NewPrefixLogger("test"),
		deviceReadWriter: mockReadWriter,
	}

	t.Run("error writing file", func(t *testing.T) {
		mockReadWriter.EXPECT().WriteFile(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("unable to write file"))
		err := s.Initialize()
		require.ErrorContains(err, "unable to write file")
	})

	t.Run("successful initialization", func(t *testing.T) {
		mockReadWriter.EXPECT().WriteFile(gomock.Any(), gomock.Any(), gomock.Any()).Times(3).Return(nil)
		err := s.Initialize()
		require.NoError(err)
	})
}

func TestEnsure(t *testing.T) {
	require := require.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReadWriter := fileio.NewMockReadWriter(ctrl)

	s := &SpecManager{
		log:              log.NewPrefixLogger("test"),
		deviceReadWriter: mockReadWriter,
	}

	t.Run("error checking if file exists", func(t *testing.T) {
		errMsg := "unable to check if file exists"
		mockReadWriter.EXPECT().FileExists(gomock.Any()).Return(false, errors.New(errMsg))
		err := s.Ensure()
		require.ErrorContains(err, errMsg)
	})

	t.Run("error writing file when it does not exist", func(t *testing.T) {
		errMsg := "write failure"
		mockReadWriter.EXPECT().FileExists(gomock.Any()).Return(false, nil)
		mockReadWriter.EXPECT().WriteFile(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New(errMsg))
		err := s.Ensure()
		require.ErrorContains(err, errMsg)
	})

	t.Run("files are written when they don't exist", func(t *testing.T) {
		mockReadWriter.EXPECT().FileExists(gomock.Any()).Times(2).Return(true, nil)
		mockReadWriter.EXPECT().FileExists(gomock.Any()).Times(1).Return(false, nil)
		mockReadWriter.EXPECT().WriteFile(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil)
		err := s.Ensure()
		require.NoError(err)
	})

	t.Run("no files are written when they all exist", func(t *testing.T) {
		mockReadWriter.EXPECT().FileExists(gomock.Any()).Times(3).Return(true, nil)
		mockReadWriter.EXPECT().WriteFile(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
		err := s.Ensure()
		require.NoError(err)
	})
}

func TestRead(t *testing.T) {
	require := require.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReadWriter := fileio.NewMockReadWriter(ctrl)

	s := &SpecManager{
		log:              log.NewPrefixLogger("test"),
		deviceReadWriter: mockReadWriter,
	}

	t.Run("error file not found", func(t *testing.T) {
		mockReadWriter.EXPECT().ReadFile(gomock.Any()).Return(nil, os.ErrNotExist)

		_, err := s.Read("current")
		require.ErrorIs(err, ErrMissingRenderedSpec)
	})

	t.Run("error with file read", func(t *testing.T) {
		errMsg := "error reading file"
		mockReadWriter.EXPECT().ReadFile(gomock.Any()).Return(nil, errors.New(errMsg))
		_, err := s.Read("current")
		require.ErrorContains(err, errMsg)
	})

	t.Run("error when the file read cannot be unmarshaled into a device spec", func(t *testing.T) {
		invalidSpec := []byte("Not json data")
		mockReadWriter.EXPECT().ReadFile(gomock.Any()).Return(invalidSpec, nil)

		_, err := s.Read("current")
		require.ErrorContains(err, "unmarshal device specification:")
	})

	t.Run("reads a device spec", func(t *testing.T) {
		image := "flightctl-device:v1"
		spec, err := createTestSpec(image)
		require.NoError(err)
		mockReadWriter.EXPECT().ReadFile(gomock.Any()).Return(spec, nil)

		specFromRead, err := s.Read("current")
		require.NoError(err)
		require.Equal(image, specFromRead.Os.Image)
	})
}

func TestUpgrade(t *testing.T) {
	require := require.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReadWriter := fileio.NewMockReadWriter(ctrl)

	desiredPath := "test/desired.json"
	currentPath := "test/current.json"
	rollbackPath := "test/rollback/json"
	s := &SpecManager{
		log:              log.NewPrefixLogger("test"),
		deviceReadWriter: mockReadWriter,
		desiredPath:      desiredPath,
		currentPath:      currentPath,
		rollbackPath:     rollbackPath,
	}

	t.Run("error reading desired spec", func(t *testing.T) {
		readErr := errors.New("unable to read file")
		mockReadWriter.EXPECT().ReadFile(desiredPath).Return(nil, readErr)

		err := s.Upgrade()
		require.ErrorIs(err, readErr)
	})

	t.Run("error writing desired spec to current", func(t *testing.T) {
		desiredSpec, err := createTestSpec("flightctl-device:v2")
		require.NoError(err)
		mockReadWriter.EXPECT().ReadFile(desiredPath).Return(desiredSpec, nil)

		writeErr := errors.New("failure writing file")
		mockReadWriter.EXPECT().WriteFile(currentPath, desiredSpec, gomock.Any()).Return(writeErr)

		err = s.Upgrade()
		require.ErrorIs(err, writeErr)
	})

	t.Run("clears out the rollback spec", func(t *testing.T) {
		desiredSpec, err := createTestSpec("flightctl-device:v2")
		require.NoError(err)
		mockReadWriter.EXPECT().ReadFile(desiredPath).Return(desiredSpec, nil)
		mockReadWriter.EXPECT().WriteFile(currentPath, desiredSpec, gomock.Any()).Return(nil)

		emptySpec, err := json.Marshal(&v1alpha1.RenderedDeviceSpec{})
		require.NoError(err)

		mockReadWriter.EXPECT().WriteFile(rollbackPath, emptySpec, gomock.Any()).Return(nil)
		err = s.Upgrade()
		require.NoError(err)
	})
}

func TestIsOSUpdate(t *testing.T) {
	require := require.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReadWriter := fileio.NewMockReadWriter(ctrl)

	desiredPath := "test/desired.json"
	currentPath := "test/current.json"
	s := &SpecManager{
		log:              log.NewPrefixLogger("test"),
		deviceReadWriter: mockReadWriter,
		desiredPath:      desiredPath,
		currentPath:      currentPath,
	}

	emptySpec, err := json.Marshal(&v1alpha1.RenderedDeviceSpec{})
	require.NoError(err)

	t.Run("error reading current spec", func(t *testing.T) {
		readErr := errors.New("unable to read file")
		mockReadWriter.EXPECT().ReadFile(currentPath).Return(nil, readErr)

		_, err := s.IsOSUpdate()
		require.ErrorIs(err, readErr)
	})

	t.Run("error reading desired spec", func(t *testing.T) {
		readErr := errors.New("unable to read file")
		mockReadWriter.EXPECT().ReadFile(currentPath).Return(emptySpec, nil)
		mockReadWriter.EXPECT().ReadFile(desiredPath).Return(nil, readErr)

		_, err := s.IsOSUpdate()
		require.ErrorIs(err, readErr)
	})

	t.Run("both specs are empty", func(t *testing.T) {
		mockReadWriter.EXPECT().ReadFile(currentPath).Return(emptySpec, nil)
		mockReadWriter.EXPECT().ReadFile(desiredPath).Return(emptySpec, nil)

		osUpdate, err := s.IsOSUpdate()
		require.NoError(err)
		require.Equal(false, osUpdate)
	})

	t.Run("current and desired os images are the same", func(t *testing.T) {
		image := "flightctl-device:v2"

		currentSpec, err := createTestSpec(image)
		require.NoError(err)
		mockReadWriter.EXPECT().ReadFile(currentPath).Return(currentSpec, nil)

		desiredSpec, err := createTestSpec(image)
		require.NoError(err)
		mockReadWriter.EXPECT().ReadFile(desiredPath).Return(desiredSpec, nil)

		osUpdate, err := s.IsOSUpdate()
		require.NoError(err)
		require.Equal(false, osUpdate)
	})

	t.Run("current and desired os images are different", func(t *testing.T) {
		currentImage := "flightctl-device:v2"
		desiredImage := "flightctl-deivce:v3"

		currentSpec, err := createTestSpec(currentImage)
		require.NoError(err)
		mockReadWriter.EXPECT().ReadFile(currentPath).Return(currentSpec, nil)

		desiredSpec, err := createTestSpec(desiredImage)
		require.NoError(err)
		mockReadWriter.EXPECT().ReadFile(desiredPath).Return(desiredSpec, nil)

		osUpdate, err := s.IsOSUpdate()
		require.NoError(err)
		require.Equal(true, osUpdate)
	})
}

func TestCheckOsReconciliation(t *testing.T) {
	require := require.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReadWriter := fileio.NewMockReadWriter(ctrl)
	mockBootcClient := container.NewMockBootcClient(ctrl)

	desiredPath := "test/desired.json"
	s := &SpecManager{
		log:              log.NewPrefixLogger("test"),
		deviceReadWriter: mockReadWriter,
		bootcClient:      mockBootcClient,
		desiredPath:      desiredPath,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	emptySpec, err := json.Marshal(&v1alpha1.RenderedDeviceSpec{})
	require.NoError(err)

	t.Run("error getting bootc status", func(t *testing.T) {
		bootcErr := errors.New("bootc problem")
		mockBootcClient.EXPECT().Status(ctx).Return(nil, bootcErr)

		_, _, err := s.CheckOsReconciliation(ctx)
		// TODO consolodate / have more consistency with how we check errors in the various tests, possibly
		// worth doing with the audit of messages
		require.ErrorContains(err, "getting current bootc status: bootc problem")
	})

	// TODO should these different test cases all have these failure checks for stuff like the read?
	// could it be consolidated almost like a dependency or shared test case for the various functions?
	//
	// e.g. maybe the testing of paths should be taken into the Read test then ignored in other test cases since the methods
	// being tested are reliant on the Read behavior and don't really care beyond it working
	t.Run("error reading desired spec", func(t *testing.T) {
		bootcStatus := &container.BootcHost{}
		mockBootcClient.EXPECT().Status(ctx).Return(bootcStatus, nil)

		readErr := errors.New("unable to read file")
		mockReadWriter.EXPECT().ReadFile(desiredPath).Return(emptySpec, readErr)

		_, _, err = s.CheckOsReconciliation(ctx)
		require.ErrorIs(err, readErr)
	})

	t.Run("desired os is not set in the spec", func(t *testing.T) {
		bootedImage := "flightctl-device:v1"

		bootcStatus := &container.BootcHost{}
		bootcStatus.Status.Booted.Image.Image.Image = bootedImage
		mockBootcClient.EXPECT().Status(ctx).Return(bootcStatus, nil)

		mockReadWriter.EXPECT().ReadFile(desiredPath).Return(emptySpec, nil)

		bootedOSImage, desiredImageIsBooted, err := s.CheckOsReconciliation(ctx)
		require.NoError(err)
		require.Equal(bootedOSImage, bootedImage)
		require.Equal(false, desiredImageIsBooted)
	})

	t.Run("booted image and desired image are different", func(t *testing.T) {
		bootedImage := "flightctl-device:v1"
		desiredImage := "flightctl-device:v2"

		bootcStatus := &container.BootcHost{}
		bootcStatus.Status.Booted.Image.Image.Image = bootedImage
		mockBootcClient.EXPECT().Status(ctx).Return(bootcStatus, nil)

		desiredSpec, err := createTestSpec(desiredImage)
		require.NoError(err)
		mockReadWriter.EXPECT().ReadFile(desiredPath).Return(desiredSpec, nil)

		bootedOSImage, desiredImageIsBooted, err := s.CheckOsReconciliation(ctx)
		require.NoError(err)
		require.Equal(bootedOSImage, bootedImage)
		require.Equal(false, desiredImageIsBooted)
	})

	t.Run("booted image and desired image are the same", func(t *testing.T) {
		image := "flightctl-device:v2"

		bootcStatus := &container.BootcHost{}
		bootcStatus.Status.Booted.Image.Image.Image = image
		mockBootcClient.EXPECT().Status(ctx).Return(bootcStatus, nil)

		desiredSpec, err := createTestSpec(image)
		require.NoError(err)
		mockReadWriter.EXPECT().ReadFile(desiredPath).Return(desiredSpec, nil)

		bootedOSImage, desiredImageIsBooted, err := s.CheckOsReconciliation(ctx)
		require.NoError(err)
		require.Equal(bootedOSImage, image)
		require.Equal(true, desiredImageIsBooted)
	})
}

func TestPrepareRollback(t *testing.T) {
	require := require.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReadWriter := fileio.NewMockReadWriter(ctrl)
	mockBootcClient := container.NewMockBootcClient(ctrl)

	rollbackPath := "test/rollback.json"
	s := &SpecManager{
		log:              log.NewPrefixLogger("test"),
		deviceReadWriter: mockReadWriter,
		bootcClient:      mockBootcClient,
		rollbackPath:     rollbackPath,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	emptySpec, err := json.Marshal(&v1alpha1.RenderedDeviceSpec{})
	require.NoError(err)

	t.Run("uses the os image from the current spec in the rollback spec", func(t *testing.T) {
		currentImage := "flightctl-device:v1"

		currentSpec, err := createTestSpec(currentImage)
		require.NoError(err)
		mockReadWriter.EXPECT().ReadFile(gomock.Any()).Return(currentSpec, nil)

		rollbackSpec := &v1alpha1.RenderedDeviceSpec{
			RenderedVersion: "1",
			Os:              &v1alpha1.DeviceOSSpec{Image: currentImage},
		}
		marshaled, err := json.Marshal(rollbackSpec)
		require.NoError(err)
		mockReadWriter.EXPECT().WriteFile(rollbackPath, marshaled, gomock.Any()).Return(nil)

		err = s.PrepareRollback(ctx)
		require.NoError(err)
	})

	t.Run("falls back to the os image from bootc when the current spec os image is empty", func(t *testing.T) {
		bootedImage := "flightctl-device:v1"
		mockReadWriter.EXPECT().ReadFile(gomock.Any()).Return(emptySpec, nil)
		bootcStatus := &container.BootcHost{}
		bootcStatus.Status.Booted.Image.Image.Image = bootedImage
		mockBootcClient.EXPECT().Status(ctx).Return(bootcStatus, nil)

		rollbackSpec := &v1alpha1.RenderedDeviceSpec{
			RenderedVersion: "",
			Os:              &v1alpha1.DeviceOSSpec{Image: bootedImage},
		}
		marshaled, err := json.Marshal(rollbackSpec)
		require.NoError(err)
		mockReadWriter.EXPECT().WriteFile(rollbackPath, marshaled, gomock.Any()).Return(nil)

		err = s.PrepareRollback(ctx)
		require.NoError(err)
	})
}

func TestRollback(t *testing.T) {
	require := require.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReadWriter := fileio.NewMockReadWriter(ctrl)

	currentPath := "test/current.json"
	desiredPath := "test/desired.json"
	s := &SpecManager{
		log:              log.NewPrefixLogger("test"), // TODO do i need this logger in all the test cases?  Probably not?
		deviceReadWriter: mockReadWriter,
		currentPath:      currentPath,
		desiredPath:      desiredPath,
	}

	t.Run("error when copy fails", func(t *testing.T) {
		copyErr := errors.New("failure to copy file")
		mockReadWriter.EXPECT().CopyFile(currentPath, desiredPath).Return(copyErr)

		err := s.Rollback()
		require.ErrorIs(err, copyErr)
	})

	t.Run("copies the current spec to the desired spec", func(t *testing.T) {
		mockReadWriter.EXPECT().CopyFile(currentPath, desiredPath).Return(nil)
		err := s.Rollback()
		require.NoError(err)
	})
}

func TestSetClient(t *testing.T) {
	require := require.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := client.NewMockManagement(ctrl)

	t.Run("sets the client", func(t *testing.T) {
		s := &SpecManager{}
		s.SetClient(mockClient)
		require.Equal(mockClient, s.managementClient)
	})
}

// TODO delete
// func (s *SpecManager) GetDesired(ctx context.Context, currentRenderedVersion string) (*v1alpha1.RenderedDeviceSpec, error) {
// 	desired, err := s.Read(Desired)
// 	if err != nil {
// 		return nil, fmt.Errorf("read desired rendered spec: %w", err)
// 	}

// 	rollback, err := s.Read(Rollback)
// 	if err != nil {
// 		return nil, fmt.Errorf("read rollback rendered spec: %w", err)
// 	}

// 	renderedVersion, err := s.getRenderedVersion(currentRenderedVersion, desired.RenderedVersion, rollback.RenderedVersion)
// 	if err != nil {
// 		return nil, fmt.Errorf("get next rendered version: %w", err)
// 	}

// 	newDesired := &v1alpha1.RenderedDeviceSpec{}
// 	err = wait.ExponentialBackoff(s.backoff, func() (bool, error) {
// 		return s.getRenderedFromManagementAPIWithRetry(ctx, renderedVersion, newDesired)
// 	})
// 	if err != nil {
// 		// no content means there is no new rendered version
// 		if errors.Is(err, ErrNoContent) {
// 			s.log.Debug("No content from management API, falling back to the desired spec on disk")
// 			// TODO: can we avoid resync or is this necessary?
// 			return desired, nil
// 		}
// 		s.log.Warnf("Failed to get rendered device spec after retry: %v", err)
// 		return nil, err
// 	}

// 	s.log.Infof("Received desired rendered spec from management service with rendered version: %s", newDesired.RenderedVersion)
// 	if newDesired.RenderedVersion == desired.RenderedVersion {
// 		s.log.Infof("No new rendered version from management service, retry reconciling version: %s", newDesired.RenderedVersion)
// 		return desired, nil
// 	}

//		// write to disk
//		s.log.Infof("Writing desired rendered spec to disk with rendered version: %s", newDesired.RenderedVersion)
//		if err := s.write(Desired, newDesired); err != nil {
//			return nil, fmt.Errorf("write rendered spec to disk: %w", err)
//		}
//		return newDesired, nil
//	}
func TestGetDesired(t *testing.T) {
	require := require.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := client.NewMockManagement(ctrl)
	mockReadWriter := fileio.NewMockReadWriter(ctrl)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	desiredPath := "test/desired.json"
	rollbackPath := "test/rollback.json"
	deviceName := "test-device"
	backoff := wait.Backoff{
		Cap:      3 * time.Minute,
		Duration: 10 * time.Second,
		Factor:   1.5,
		Steps:    24,
	}
	s := &SpecManager{
		backoff:          backoff,
		log:              log.NewPrefixLogger("test"),
		deviceName:       deviceName,
		deviceReadWriter: mockReadWriter,
		desiredPath:      desiredPath,
		rollbackPath:     rollbackPath,
		managementClient: mockClient,
	}

	image := "flightctl-device:v2"

	testCases := []struct {
		Name                      string
		CurrentRenderedVersion    string
		DesiredRenderedVersion    string
		RollbackRenderedVersion   string
		NewDesiredRenderedVersion string
		NewDesiredReturnExpected  bool
	}{
		{"first", "1", "1", "1", "1", false},
		{"second", "1", "1", "1", "2", true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			desiredSpec, err := createTestSpecWithRenderedVersion(image, testCase.DesiredRenderedVersion)
			require.NoError(err)
			mockReadWriter.EXPECT().ReadFile(desiredPath).Return(desiredSpec, nil)

			rollbackSpec, err := createTestSpecWithRenderedVersion(image, testCase.RollbackRenderedVersion)
			require.NoError(err)
			mockReadWriter.EXPECT().ReadFile(rollbackPath).Return(rollbackSpec, nil)

			expectedParams := &v1alpha1.GetRenderedDeviceSpecParams{}
			// TODO this will change
			expectedParams.KnownRenderedVersion = &testCase.CurrentRenderedVersion
			apiResponse := &v1alpha1.RenderedDeviceSpec{RenderedVersion: testCase.NewDesiredRenderedVersion}
			mockClient.EXPECT().GetRenderedDeviceSpec(ctx, gomock.Any(), gomock.Any()).Return(apiResponse, 200, nil)

			if testCase.NewDesiredReturnExpected {
				mockReadWriter.EXPECT().WriteFile(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			}

			specResult, err := s.GetDesired(ctx, testCase.CurrentRenderedVersion)

			if testCase.NewDesiredReturnExpected {
				require.NoError(err)
				require.Equal(*apiResponse, *specResult)
			} else {
				require.NoError(err)
				unmarshaled := &v1alpha1.RenderedDeviceSpec{}
				err = json.Unmarshal(desiredSpec, unmarshaled)
				require.NoError(err)
				require.Equal(*unmarshaled, *specResult)
			}
		})
	}
}

func TestNewManager(t *testing.T) {
	require := require.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReadWriter := fileio.NewMockReadWriter(ctrl)
	mockBootcClient := container.NewMockBootcClient(ctrl)
	logger := log.NewPrefixLogger("test")
	backoff := createTestBackoff()

	t.Run("constructs file paths for the spec files", func(t *testing.T) {
		manager := NewManager(
			"device-name",
			"test/directory/structure/",
			mockReadWriter,
			mockBootcClient,
			backoff,
			logger,
		)

		require.Equal("test/directory/structure/current.json", manager.currentPath)
		require.Equal("test/directory/structure/desired.json", manager.desiredPath)
		require.Equal("test/directory/structure/rollback.json", manager.rollbackPath)
	})
}

func Test_getNextRenderedVersion(t *testing.T) {
	require := require.New(t)
	testCases := []struct {
		Name                string
		RenderedVersion     string
		NextRenderedVersion string
		ExpectsError        bool
	}{
		{"empty rendered version", "", "", false},
		{"increments the rendered version", "1", "2", false},
		{"errors when the rendered version cannot be parsed", "not-a-number", "", true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			nextVersion, err := getNextRenderedVersion(testCase.RenderedVersion)

			if testCase.ExpectsError {
				require.ErrorContains(err, "failed to convert version to integer:")
			} else {
				require.NoError(err)
			}

			require.Equal(testCase.NextRenderedVersion, nextVersion)
		})
	}
}

func createTestSpec(image string) ([]byte, error) {
	spec := v1alpha1.RenderedDeviceSpec{
		Os: &v1alpha1.DeviceOSSpec{
			Image: image,
		},
		RenderedVersion: "1",
	}
	return json.Marshal(spec)
}

// TODO consolidate?
func createTestSpecWithRenderedVersion(image string, renderedVersion string) ([]byte, error) {
	spec := v1alpha1.RenderedDeviceSpec{
		Os: &v1alpha1.DeviceOSSpec{
			Image: image,
		},
		RenderedVersion: renderedVersion,
	}
	return json.Marshal(spec)
}

// TODO change these values
func createTestBackoff() wait.Backoff {
	return wait.Backoff{
		Cap:      3 * time.Minute,
		Duration: 10 * time.Second,
		Factor:   1.5,
		Steps:    24,
	}
}
