package spec

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/flightctl/flightctl/api/v1alpha1"
	"github.com/flightctl/flightctl/internal/agent/device/fileio"
	"github.com/flightctl/flightctl/internal/container"
	"github.com/flightctl/flightctl/pkg/log"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
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

	// emptySpec, err := json.Marshal(&v1alpha1.RenderedDeviceSpec{})
	// require.NoError(err)

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
