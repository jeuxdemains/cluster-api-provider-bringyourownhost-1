// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package installer

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type mockImgpkg struct {
	callCount int
	err       string
}

func (mi *mockImgpkg) Get(_, _ string) error {
	mi.callCount++
	return nil
}

func (mi *mockImgpkg) GetError(_, _ string) error {
	return errors.New(mi.err)
}

var _ = Describe("Byohost Installer Tests", func() {

	var (
		bd                  *bundleDownloader
		mi                  *mockImgpkg
		repoAddr            string
		downloadPath        string
		normalizedOsVersion string
		k8sVersion          string
	)

	BeforeEach(func() {
		normalizedOsVersion = "Ubuntu_20.04.3_x64"
		k8sVersion = "1.22"
		repoAddr = ""
		var err error
		downloadPath, err = os.MkdirTemp("", "downloaderTest")
		if err != nil {
			log.Fatal(err)
		}
		bd = &bundleDownloader{repoAddr, downloadPath}
		mi = &mockImgpkg{}
	})
	AfterEach(func() {
		err := os.RemoveAll(downloadPath)
		if err != nil {
			log.Fatal(err)
		}
	})
	Context("When given correct arguments", func() {

		It("Should download bundle", func() {
			// Test download on cache missing
			err := bd.DownloadFromRepo(
				normalizedOsVersion,
				k8sVersion,
				func(a, b string) error { return mi.Get(a, b) })
			Expect(err).ShouldNot((HaveOccurred()))

			// Test no download on cache hit
			err = bd.DownloadFromRepo(
				normalizedOsVersion,
				k8sVersion,
				func(a, b string) error { return mi.Get(a, b) })
			Expect(err).ShouldNot((HaveOccurred()))
			Expect(mi.callCount).Should(Equal(1))
		})
		It("Should create dir if missing and download bundle", func() {
			bd.downloadPath = filepath.Join(bd.downloadPath, "a", "b", "c")
			err := bd.DownloadFromRepo(
				normalizedOsVersion,
				k8sVersion,
				func(a, b string) error { return mi.Get(a, b) })
			time.Sleep(8 * time.Second)
			Expect(err).ShouldNot((HaveOccurred()))
		})
	})
	Context("When there is error during download", func() {
		It("Should return error if given bad repo", func() {
			bd.repoAddr = "a.a"
			err := bd.Download(normalizedOsVersion, k8sVersion)
			Expect(err).Should((HaveOccurred()))
			Expect(err.Error()).Should(Equal(ErrBundleDownload.Error()))
		})
		It("Should return error if connection timed out", func() {
			mi.err = "Extracting image into directory: read tcp 192.168.0.1:1->1.1.1.1:1: read: connection timed out"
			err := bd.DownloadFromRepo(
				normalizedOsVersion,
				k8sVersion,
				func(a, b string) error { return mi.GetError(a, b) })
			Expect(err).Should((HaveOccurred()))
			Expect(err.Error()).Should(Equal(ErrBundleDownload.Error()))
		})
		It("Should return error if failure in name resolution", func() {
			mi.err = "Fetching image: Get \"a.a/\": dial tcp: lookup a.a: Temporary failure in name resolution"
			err := bd.DownloadFromRepo(
				normalizedOsVersion,
				k8sVersion,
				func(a, b string) error { return mi.GetError(a, b) })
			Expect(err).Should((HaveOccurred()))
			Expect(err.Error()).Should(Equal(ErrBundleDownload.Error()))
		})
		It("Should return error if out of space", func() {
			mi.err = "Extracting image into directory: write /tmp/asd: no space left on device"
			err := bd.DownloadFromRepo(
				normalizedOsVersion,
				k8sVersion,
				func(a, b string) error { return mi.GetError(a, b) })
			Expect(err).Should((HaveOccurred()))
			Expect(err.Error()).Should(Equal(ErrBundleExtract.Error()))
		})

	})
})