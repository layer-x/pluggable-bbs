package etcd_test

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/cloudfoundry-incubator/auctioneer"
	"github.com/cloudfoundry-incubator/bbs/db/etcd"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/bbs/models/test/model_helpers"
	etcdclient "github.com/coreos/go-etcd/etcd"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DesiredLRPDB", func() {
	Describe("DesiredLRPs", func() {
		var filter models.DesiredLRPFilter
		var desiredLRPsInDomains map[string][]*models.DesiredLRP

		BeforeEach(func() {
			filter = models.DesiredLRPFilter{}
		})

		Context("when there are desired LRPs", func() {
			var expectedDesiredLRPs []*models.DesiredLRP

			BeforeEach(func() {
				expectedDesiredLRPs = []*models.DesiredLRP{}

				desiredLRPsInDomains = etcdHelper.CreateDesiredLRPsInDomains(map[string]int{
					"domain-1": 1,
					"domain-2": 2,
				})
			})

			It("returns all the desired LRPs", func() {
				for _, domainLRPs := range desiredLRPsInDomains {
					for _, lrp := range domainLRPs {
						expectedDesiredLRPs = append(expectedDesiredLRPs, lrp)
					}
				}
				desiredLRPs, err := etcdDB.DesiredLRPs(logger, filter)
				Expect(err).NotTo(HaveOccurred())
				Expect(desiredLRPs).To(ConsistOf(expectedDesiredLRPs))
			})

			It("can filter by domain", func() {
				for _, lrp := range desiredLRPsInDomains["domain-2"] {
					expectedDesiredLRPs = append(expectedDesiredLRPs, lrp)
				}
				filter.Domain = "domain-2"
				desiredLRPs, err := etcdDB.DesiredLRPs(logger, filter)
				Expect(err).NotTo(HaveOccurred())
				Expect(desiredLRPs).To(ConsistOf(expectedDesiredLRPs))
			})
		})

		Context("when there are no LRPs", func() {
			It("returns an empty list", func() {
				desiredLRPs, err := etcdDB.DesiredLRPs(logger, filter)
				Expect(err).NotTo(HaveOccurred())
				Expect(desiredLRPs).NotTo(BeNil())
				Expect(desiredLRPs).To(BeEmpty())
			})
		})

		Context("when there is invalid data", func() {
			BeforeEach(func() {
				etcdHelper.CreateValidDesiredLRP("guid-1")
				etcdHelper.CreateMalformedDesiredLRP("bad-guid")
				etcdHelper.CreateValidDesiredLRP("guid-2")
			})

			It("retuns only valid records", func() {
				desireds, err := etcdDB.DesiredLRPs(logger, filter)
				Expect(err).ToNot(HaveOccurred())
				Expect(desireds).To(HaveLen(2))
				Expect([]string{desireds[0].ProcessGuid, desireds[1].ProcessGuid}).To(ConsistOf("guid-1", "guid-2"))
			})
		})

		Context("when etcd is not there", func() {
			BeforeEach(func() {
				etcdRunner.Stop()
			})

			AfterEach(func() {
				etcdRunner.Start()
			})

			It("errors", func() {
				_, err := etcdDB.DesiredLRPs(logger, filter)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("DesiredLRPByProcessGuid", func() {
		Context("when there is a desired lrp", func() {
			var desiredLRP *models.DesiredLRP

			BeforeEach(func() {
				desiredLRP = model_helpers.NewValidDesiredLRP("process-guid")
				etcdHelper.SetRawDesiredLRP(desiredLRP)
			})

			It("returns the desired lrp", func() {
				lrp, err := etcdDB.DesiredLRPByProcessGuid(logger, "process-guid")
				Expect(err).NotTo(HaveOccurred())
				Expect(lrp).To(Equal(desiredLRP))
			})
		})

		Context("when there is no LRP", func() {
			It("returns a ResourceNotFound", func() {
				_, err := etcdDB.DesiredLRPByProcessGuid(logger, "nota-guid")
				Expect(err).To(Equal(models.ErrResourceNotFound))
			})
		})

		Context("when there is invalid data", func() {
			BeforeEach(func() {
				etcdHelper.CreateMalformedDesiredLRP("some-other-guid")
			})

			It("errors", func() {
				_, err := etcdDB.DesiredLRPByProcessGuid(logger, "some-other-guid")
				Expect(err).To(HaveOccurred())
				bbsErr := models.ConvertError(err)
				Expect(bbsErr.Type).To(Equal(models.Error_InvalidRecord))
			})
		})

		Context("when etcd is not there", func() {
			BeforeEach(func() {
				etcdRunner.Stop()
			})

			AfterEach(func() {
				etcdRunner.Start()
			})

			It("errors", func() {
				_, err := etcdDB.DesiredLRPByProcessGuid(logger, "some-other-guid")
				Expect(err).To(Equal(models.ErrUnknownError))
			})
		})
	})

	Describe("DesiredLRPSchedulingInfos", func() {
		var filter models.DesiredLRPFilter
		var desiredLRPsInDomains map[string][]*models.DesiredLRP

		BeforeEach(func() {
			filter = models.DesiredLRPFilter{}
		})

		Context("when there are desired LRPs", func() {
			var expectedSchedulingInfos []*models.DesiredLRPSchedulingInfo

			BeforeEach(func() {
				expectedSchedulingInfos = []*models.DesiredLRPSchedulingInfo{}

				desiredLRPsInDomains = etcdHelper.CreateDesiredLRPsInDomains(map[string]int{
					"domain-1": 1,
					"domain-2": 2,
				})
			})

			It("returns all the scheduling infos", func() {
				for _, domainLRPs := range desiredLRPsInDomains {
					for _, lrp := range domainLRPs {
						schedulingInfo := lrp.DesiredLRPSchedulingInfo()
						expectedSchedulingInfos = append(expectedSchedulingInfos, &schedulingInfo)
					}
				}
				schedulingInfos, err := etcdDB.DesiredLRPSchedulingInfos(logger, filter)
				Expect(err).NotTo(HaveOccurred())
				Expect(schedulingInfos).To(ConsistOf(expectedSchedulingInfos))
			})

			It("can filter by domain", func() {
				for _, lrp := range desiredLRPsInDomains["domain-2"] {
					schedulingInfo := lrp.DesiredLRPSchedulingInfo()
					expectedSchedulingInfos = append(expectedSchedulingInfos, &schedulingInfo)
				}
				filter.Domain = "domain-2"
				schedulingInfos, err := etcdDB.DesiredLRPSchedulingInfos(logger, filter)
				Expect(err).NotTo(HaveOccurred())
				Expect(schedulingInfos).To(ConsistOf(expectedSchedulingInfos))
			})
		})

		Context("when there are no LRPs", func() {
			It("returns an empty list", func() {
				schedulingInfos, err := etcdDB.DesiredLRPSchedulingInfos(logger, filter)
				Expect(err).NotTo(HaveOccurred())
				Expect(schedulingInfos).NotTo(BeNil())
				Expect(schedulingInfos).To(BeEmpty())
			})
		})

		Context("when there is invalid data", func() {
			BeforeEach(func() {
				etcdHelper.CreateValidDesiredLRP("guid-1")
				etcdHelper.CreateMalformedDesiredLRP("bad-guid")
				etcdHelper.CreateValidDesiredLRP("guid-2")
			})

			It("retuns only valid records", func() {
				schedulingInfo, err := etcdDB.DesiredLRPSchedulingInfos(logger, filter)
				Expect(err).ToNot(HaveOccurred())
				Expect(schedulingInfo).To(HaveLen(2))
				Expect([]string{schedulingInfo[0].ProcessGuid, schedulingInfo[1].ProcessGuid}).To(ConsistOf("guid-1", "guid-2"))
			})
		})

		Context("when etcd is not there", func() {
			BeforeEach(func() {
				etcdRunner.Stop()
			})

			AfterEach(func() {
				etcdRunner.Start()
			})

			It("errors", func() {
				_, err := etcdDB.DesiredLRPSchedulingInfos(logger, filter)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("DesireLRP", func() {
		var lrp *models.DesiredLRP

		BeforeEach(func() {
			lrp = model_helpers.NewValidDesiredLRP("some-process-guid")
			lrp.Instances = 5
		})

		Context("when the desired LRP does not yet exist", func() {
			It("persists the scheduling info and run info", func() {
				err := etcdDB.DesireLRP(logger, lrp)
				Expect(err).NotTo(HaveOccurred())

				persisted, err := etcdDB.DesiredLRPByProcessGuid(logger, "some-process-guid")
				Expect(err).NotTo(HaveOccurred())

				Expect(persisted.DesiredLRPKey()).To(Equal(lrp.DesiredLRPKey()))
				Expect(persisted.DesiredLRPResource()).To(Equal(lrp.DesiredLRPResource()))
				Expect(persisted.Annotation).To(Equal(lrp.Annotation))
				Expect(persisted.Instances).To(Equal(lrp.Instances))
				Expect(persisted.DesiredLRPRunInfo(clock.Now())).To(Equal(lrp.DesiredLRPRunInfo(clock.Now())))
			})

			It("creates one ActualLRP per index", func() {
				err := etcdDB.DesireLRP(logger, lrp)
				Expect(err).NotTo(HaveOccurred())
				actualLRPGroups, err := etcdDB.ActualLRPGroupsByProcessGuid(logger, "some-process-guid")
				Expect(err).NotTo(HaveOccurred())
				Expect(actualLRPGroups).To(HaveLen(5))
			})

			It("sets a ModificationTag on each ActualLRP with a unique epoch", func() {
				err := etcdDB.DesireLRP(logger, lrp)
				Expect(err).NotTo(HaveOccurred())
				actualLRPGroups, err := etcdDB.ActualLRPGroupsByProcessGuid(logger, "some-process-guid")
				Expect(err).NotTo(HaveOccurred())

				epochs := map[string]models.ActualLRP{}
				for _, actualLRPGroup := range actualLRPGroups {
					epochs[actualLRPGroup.Instance.ModificationTag.Epoch] = *actualLRPGroup.Instance
				}

				Expect(epochs).To(HaveLen(5))
			})

			It("sets the ModificationTag on the DesiredLRP", func() {
				err := etcdDB.DesireLRP(logger, lrp)
				Expect(err).NotTo(HaveOccurred())

				lrp, err := etcdDB.DesiredLRPByProcessGuid(logger, "some-process-guid")
				Expect(err).NotTo(HaveOccurred())

				Expect(lrp.ModificationTag.Epoch).NotTo(BeEmpty())
				Expect(lrp.ModificationTag.Index).To(BeEquivalentTo(0))
			})

			Context("when an auctioneer is present", func() {
				It("emits start auction requests", func() {
					originalAuctionCallCount := fakeAuctioneerClient.RequestLRPAuctionsCallCount()

					err := etcdDB.DesireLRP(logger, lrp)
					Expect(err).NotTo(HaveOccurred())

					desired, err := etcdDB.DesiredLRPByProcessGuid(logger, lrp.ProcessGuid)
					Expect(err).NotTo(HaveOccurred())

					Consistently(fakeAuctioneerClient.RequestLRPAuctionsCallCount).Should(Equal(originalAuctionCallCount + 1))

					expectedStartRequest := auctioneer.NewLRPStartRequestFromModel(desired, 0, 1, 2, 3, 4)

					startAuctions := fakeAuctioneerClient.RequestLRPAuctionsArgsForCall(originalAuctionCallCount)
					Expect(startAuctions).To(HaveLen(1))
					Expect(startAuctions[0].ProcessGuid).To(Equal(desired.ProcessGuid))
					Expect(startAuctions[0].Indices).To(ConsistOf(expectedStartRequest.Indices))
				})
			})

			Context("An error occurs creating the scheduling info", func() {
				BeforeEach(func() {
					count := 0
					fakeStoreClient.CreateStub = func(key string, value []byte, ttl uint64) (*etcdclient.Response, error) {
						if count == 0 {
							count++
							return nil, nil
						} else {
							return nil, errors.New("Failed Scheduling desired lrp ingo")
						}
					}
				})

				It("attempts to delete the run info", func() {
					err := etcdDBWithFakeStore.DesireLRP(logger, lrp)
					Expect(err).To(HaveOccurred())

					Expect(fakeStoreClient.DeleteCallCount()).To(Equal(1))
					schemaPath, _ := fakeStoreClient.DeleteArgsForCall(0)
					Expect(schemaPath).To(Equal(etcd.DesiredLRPRunInfoSchemaPath(lrp.ProcessGuid)))
				})
			})
		})

		Context("when the desired LRP already exists", func() {
			var newLRP *models.DesiredLRP

			BeforeEach(func() {
				err := etcdDB.DesireLRP(logger, lrp)
				Expect(err).NotTo(HaveOccurred())

				newLRP = lrp
				newLRP.Instances = 3
			})

			It("rejects the request with ErrResourceExists", func() {
				err := etcdDB.DesireLRP(logger, newLRP)
				Expect(err).To(Equal(models.ErrResourceExists))
			})
		})
	})

	Describe("RemoveDesiredLRP", func() {
		var lrp *models.DesiredLRP

		BeforeEach(func() {
			lrp = model_helpers.NewValidDesiredLRP("some-process-guid")
			lrp.Instances = 5
		})

		Context("when the desired LRP exists", func() {
			BeforeEach(func() {
				err := etcdDB.DesireLRP(logger, lrp)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should delete it", func() {
				err := etcdDB.RemoveDesiredLRP(logger, lrp.ProcessGuid)
				Expect(err).NotTo(HaveOccurred())

				_, err = etcdDB.DesiredLRPByProcessGuid(logger, lrp.ProcessGuid)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(models.ErrResourceNotFound))
			})

			Context("when there are running instances on a present cell", func() {
				cellPresence := models.NewCellPresence("the-cell-id", "cell.example.com", "az1", models.NewCellCapacity(128, 1024, 6), []string{}, []string{})

				BeforeEach(func() {
					consulHelper.RegisterCell(&cellPresence)

					for i := int32(0); i < lrp.Instances; i++ {
						instanceKey := models.NewActualLRPInstanceKey(fmt.Sprintf("some-instance-guid-%d", i), cellPresence.CellID)
						err := etcdDB.ClaimActualLRP(logger, lrp.ProcessGuid, i, &instanceKey)
						Expect(err).NotTo(HaveOccurred())
					}
				})

				It("stops all actual lrps for the desired lrp", func() {
					originalStopCallCount := fakeRepClient.StopLRPInstanceCallCount()

					err := etcdDB.RemoveDesiredLRP(logger, lrp.ProcessGuid)
					Expect(err).NotTo(HaveOccurred())

					callCount := originalStopCallCount + int(lrp.Instances)

					Expect(fakeRepClientFactory.CreateClientCallCount()).To(Equal(callCount))
					Expect(fakeRepClientFactory.CreateClientArgsForCall(0)).To(Equal(cellPresence.RepAddress))

					Expect(fakeRepClient.StopLRPInstanceCallCount()).To(Equal(callCount))

					stoppedActuals := make([]int32, lrp.Instances)
					for i := int32(0); i < lrp.Instances; i++ {
						key, _ := fakeRepClient.StopLRPInstanceArgsForCall(originalStopCallCount + int(i))
						stoppedActuals[i] = key.Index
					}
					Expect(stoppedActuals).To(ConsistOf([]int32{0, 1, 2, 3, 4}))
				})
			})
		})

		Context("when the RunInfo exists, and the SchedulingInfo does not exist", func() {
			BeforeEach(func() {
				err := etcdDB.DesireLRP(logger, lrp)
				Expect(err).NotTo(HaveOccurred())
				_, err = storeClient.Delete(etcd.DesiredLRPSchedulingInfoSchemaPath(lrp.ProcessGuid), true)
				Expect(err).NotTo(HaveOccurred())
			})

			It("deletes the RunInfo", func() {
				err := etcdDB.RemoveDesiredLRP(logger, lrp.ProcessGuid)
				Expect(err).ToNot(HaveOccurred())
				_, err = etcdDB.DesiredLRPByProcessGuid(logger, lrp.ProcessGuid)
				Expect(err).To(Equal(models.ErrResourceNotFound))
				_, err = storeClient.Get(etcd.DesiredLRPRunInfoSchemaPath(lrp.ProcessGuid), false, false)
				Expect(etcd.ErrorFromEtcdError(logger, err)).To(Equal(models.ErrResourceNotFound))
			})
		})

		Context("when removing the SchedulingInfo fails", func() {
			BeforeEach(func() {
				fakeStoreClient.DeleteReturns(nil, errors.New("kabooom!"))

				err := etcdDBWithFakeStore.DesireLRP(logger, lrp)
				Expect(err).NotTo(HaveOccurred())
			})

			It("does not remove the RunInfo", func() {
				err := etcdDBWithFakeStore.RemoveDesiredLRP(logger, lrp.ProcessGuid)
				Expect(err).To(HaveOccurred())

				Expect(fakeStoreClient.DeleteCallCount()).To(Equal(1))
				schemaPath, _ := fakeStoreClient.DeleteArgsForCall(0)
				Expect(schemaPath).To(Equal(etcd.DesiredLRPSchedulingInfoSchemaPath(lrp.ProcessGuid)))
			})
		})

		Context("when the desired LRP does not exist", func() {
			It("returns a resource not found error", func() {
				err := etcdDB.RemoveDesiredLRP(logger, "monkey")
				Expect(err).To(Equal(models.ErrResourceNotFound))
			})
		})
	})

	Describe("Updating DesireLRP", func() {
		var (
			update     *models.DesiredLRPUpdate
			desiredLRP *models.DesiredLRP
			lrp        *models.DesiredLRP
		)

		BeforeEach(func() {
			lrp = model_helpers.NewValidDesiredLRP("some-process-guid")
			lrp.Instances = 5
			err := etcdDB.DesireLRP(logger, lrp)
			Expect(err).NotTo(HaveOccurred())

			desiredLRP, err = etcdDB.DesiredLRPByProcessGuid(logger, lrp.ProcessGuid)
			Expect(err).NotTo(HaveOccurred())

			update = &models.DesiredLRPUpdate{}
		})

		Context("When the updates are valid", func() {
			BeforeEach(func() {
				annotation := "new-annotation"
				instances := int32(16)

				rawMessage := json.RawMessage([]byte(`{"port":8080,"hosts":["new-route-1","new-route-2"]}`))
				update.Routes = &models.Routes{
					"router": &rawMessage,
				}
				update.Annotation = &annotation
				update.Instances = &instances
			})

			It("updates an existing DesireLRP", func() {
				modelErr := etcdDB.UpdateDesiredLRP(logger, lrp.ProcessGuid, update)
				Expect(modelErr).NotTo(HaveOccurred())

				updated, modelErr := etcdDB.DesiredLRPByProcessGuid(logger, lrp.ProcessGuid)
				Expect(modelErr).NotTo(HaveOccurred())

				Expect(*updated.Routes).To(HaveKey("router"))
				json, err := (*update.Routes)["router"].MarshalJSON()
				Expect(err).NotTo(HaveOccurred())
				updatedJson, err := (*updated.Routes)["router"].MarshalJSON()
				Expect(err).NotTo(HaveOccurred())
				Expect(updatedJson).To(MatchJSON(string(json)))
				Expect(updated.Annotation).To(Equal(*update.Annotation))
				Expect(updated.Instances).To(Equal(*update.Instances))
				Expect(updated.ModificationTag.Epoch).To(Equal(desiredLRP.ModificationTag.Epoch))
				Expect(updated.ModificationTag.Index).To(Equal(desiredLRP.ModificationTag.Index + 1))
			})

			Context("when the compare and swap fails", func() {
				BeforeEach(func() {
					resp, err := storeClient.Get(etcd.DesiredLRPSchedulingInfoSchemaPath(lrp.ProcessGuid), false, false)
					Expect(err).NotTo(HaveOccurred())
					fakeStoreClient.GetReturns(resp, nil) // return the pre-updated desired lrps
				})

				Context("for a CAS failure", func() {
					BeforeEach(func() {
						fakeStoreClient.CompareAndSwapReturns(nil, etcdclient.EtcdError{ErrorCode: etcd.ETCDErrIndexComparisonFailed})
					})

					It("retries the update up to 2 times", func() {
						Expect(fakeStoreClient.CompareAndSwapCallCount()).To(Equal(0))
						modelErr := etcdDBWithFakeStore.UpdateDesiredLRP(logger, lrp.ProcessGuid, update)
						Expect(modelErr).To(HaveOccurred())
						Expect(fakeStoreClient.CompareAndSwapCallCount()).To(Equal(2))
					})
				})

				Context("for a non CAS failure", func() {
					BeforeEach(func() {
						fakeStoreClient.CompareAndSwapReturns(nil, etcdclient.EtcdError{ErrorCode: etcd.ETCDErrKeyExists})
					})

					It("fails immediately", func() {
						Expect(fakeStoreClient.CompareAndSwapCallCount()).To(Equal(0))
						modelErr := etcdDBWithFakeStore.UpdateDesiredLRP(logger, lrp.ProcessGuid, update)
						Expect(modelErr).To(HaveOccurred())
						Expect(fakeStoreClient.CompareAndSwapCallCount()).To(Equal(1))
					})
				})
			})

			Context("when the instances are increased", func() {
				BeforeEach(func() {
					instances := int32(6)
					update.Instances = &instances
				})

				Context("when an auctioneer is present", func() {
					It("emits start auction requests", func() {
						originalAuctionCallCount := fakeAuctioneerClient.RequestLRPAuctionsCallCount()

						err := etcdDB.UpdateDesiredLRP(logger, lrp.ProcessGuid, update)
						Expect(err).NotTo(HaveOccurred())

						Consistently(fakeAuctioneerClient.RequestLRPAuctionsCallCount).Should(Equal(originalAuctionCallCount + 1))

						updated, err := etcdDB.DesiredLRPByProcessGuid(logger, lrp.ProcessGuid)
						Expect(err).NotTo(HaveOccurred())

						expectedStartRequest := auctioneer.NewLRPStartRequestFromModel(updated, 5)
						startAuctions := fakeAuctioneerClient.RequestLRPAuctionsArgsForCall(originalAuctionCallCount)
						Expect(startAuctions).To(HaveLen(1))
						Expect(*startAuctions[0]).To(Equal(expectedStartRequest))
					})
				})
			})

			Context("when the instances are decreased", func() {
				BeforeEach(func() {
					instances := int32(2)
					update.Instances = &instances
				})

				Context("when the cell is present", func() {
					cellPresence := models.NewCellPresence("the-cell-id", "cell.example.com", "az1", models.NewCellCapacity(128, 1024, 6), []string{}, []string{})

					BeforeEach(func() {
						consulHelper.RegisterCell(&cellPresence)

						for i := int32(0); i < lrp.Instances; i++ {
							instanceKey := models.NewActualLRPInstanceKey(fmt.Sprintf("some-instance-guid-%d", i), cellPresence.CellID)
							err := etcdDB.ClaimActualLRP(logger, lrp.ProcessGuid, i, &instanceKey)
							Expect(err).NotTo(HaveOccurred())
						}
					})

					It("stops the instances at the removed indices", func() {
						originalStopCallCount := fakeRepClient.StopLRPInstanceCallCount()

						err := etcdDB.UpdateDesiredLRP(logger, lrp.ProcessGuid, update)
						Expect(err).NotTo(HaveOccurred())

						callCount := originalStopCallCount + int(lrp.Instances-*(update.Instances))

						Expect(fakeRepClientFactory.CreateClientCallCount()).To(Equal(callCount))
						Expect(fakeRepClientFactory.CreateClientArgsForCall(0)).To(Equal(cellPresence.RepAddress))

						Expect(fakeRepClient.StopLRPInstanceCallCount()).To(Equal(callCount))
						stoppedActuals := make([]int32, lrp.Instances-*update.Instances)
						for i := int32(0); i < (lrp.Instances - *update.Instances); i++ {
							key, _ := fakeRepClient.StopLRPInstanceArgsForCall(originalStopCallCount + int(i))
							stoppedActuals[i] = key.Index
						}

						Expect(stoppedActuals).To(ConsistOf([]int32{2, 3, 4}))
					})
				})
			})
		})

		Context("When the updates are invalid", func() {
			It("instances cannot be less than zero", func() {
				instances := int32(-1)

				update := &models.DesiredLRPUpdate{
					Instances: &instances,
				}

				desiredBeforeUpdate, err := etcdDB.DesiredLRPByProcessGuid(logger, lrp.ProcessGuid)
				Expect(err).NotTo(HaveOccurred())

				err = etcdDB.UpdateDesiredLRP(logger, lrp.ProcessGuid, update)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("instances"))

				desiredAfterUpdate, err := etcdDB.DesiredLRPByProcessGuid(logger, lrp.ProcessGuid)
				Expect(err).NotTo(HaveOccurred())

				Expect(desiredAfterUpdate).To(Equal(desiredBeforeUpdate))
			})
		})

		Context("When the LRP does not exist", func() {
			It("returns an ErrorKeyNotFound", func() {
				instances := int32(0)

				err := etcdDB.UpdateDesiredLRP(logger, "garbage-guid", &models.DesiredLRPUpdate{
					Instances: &instances,
				})
				Expect(err).To(Equal(models.ErrResourceNotFound))
			})
		})
	})
})
