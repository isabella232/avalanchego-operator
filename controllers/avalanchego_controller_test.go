package controllers

import (
	"context"
	"time"

	chainv1alpha1 "github.com/ava-labs/avalanchego-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Avalanchego controller", func() {
	const (
		AvalanchegoValidatorName           = "avalanchego-test-validator"
		AvalanchegoNamespace               = "default"
		AvalanchegoValidatorDeploymentName = "test-validator"

		AvalanchegoWorkerName           = "avalanchego-test-worker"
		AvalanchegoWorkerDeploymentName = "test-worker"

		AvalanchegoKind       = "Avalanchego"
		AvalanchegoAPIVersion = "chain.avax.network/v1alpha1"

		timeout  = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("Empty bootstrapperURL, genesis and certificates", func() {
		It("Should handle new chain creation", func() {

			spec := chainv1alpha1.AvalanchegoSpec{
				Tag:            "v1.6.3",
				DeploymentName: AvalanchegoValidatorDeploymentName,
				NodeCount:      5,
				Env: []corev1.EnvVar{
					{
						Name:  "AVAGO_LOG_LEVEL",
						Value: "debug",
					},
				},
			}
			key := types.NamespacedName{
				Name:      AvalanchegoValidatorName,
				Namespace: AvalanchegoNamespace,
			}
			toCreate := &chainv1alpha1.Avalanchego{
				TypeMeta: metav1.TypeMeta{
					Kind:       AvalanchegoKind,
					APIVersion: AvalanchegoAPIVersion,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Spec: spec,
			}

			By("Creating Avalanchego chain successfully")
			Expect(k8sClient.Create(context.Background(), toCreate)).Should(Succeed())
			time.Sleep(time.Second * 5)

			Eventually(func() bool {
				f := &chainv1alpha1.Avalanchego{}
				_ = k8sClient.Get(context.Background(), key, f)
				return f.Status.Error == ""
			}, timeout, interval).Should(BeTrue())

			By("Checking if amount of services created equals nodeCount")

			fetched := &chainv1alpha1.Avalanchego{}

			Eventually(func() bool {
				_ = k8sClient.Get(context.Background(), key, fetched)
				return fetched.Spec.NodeCount == len(fetched.Status.NetworkMembersURI)
			}, timeout, interval).Should(BeTrue())

			By("Checking, if genesis was generated")

			Expect(fetched.Status.Genesis).ShouldNot(Equal(""))

			By("Deleting the scope")
			Eventually(func() error {
				f := &chainv1alpha1.Avalanchego{}
				_ = k8sClient.Get(context.Background(), key, f)
				return k8sClient.Delete(context.Background(), f)
			}, timeout, interval).Should(Succeed())

			Eventually(func() error {
				f := &chainv1alpha1.Avalanchego{}
				return k8sClient.Get(context.Background(), key, f)
			}, timeout, interval).ShouldNot(Succeed())
		})
	})

	Context("Static genesis and certificates configuration", func() {
		It("Should handle new chain creation with predefined starting block and ceritficates", func() {

			specBootstrapper := chainv1alpha1.AvalanchegoSpec{
				Tag:            "v1.6.3",
				DeploymentName: AvalanchegoValidatorDeploymentName,
				NodeCount:      1,
				Genesis:        `{"networkID":12346,"allocations":[{"ethAddr":"0xb3d82b1367d362de99ab59a658165aff520cbd4d","avaxAddr":"X-custom1g65uqn6t77p656w64023nh8nd9updzmxwd59gh","initialAmount":0,"unlockSchedule":[{"amount":10000000000000000,"locktime":1633824000}]},{"ethAddr":"0xb3d82b1367d362de99ab59a658165aff520cbd4d","avaxAddr":"X-custom18jma8ppw3nhx5r4ap8clazz0dps7rv5u9xde7p","initialAmount":300000000000000000,"unlockSchedule":[{"amount":20000000000000000},{"amount":10000000000000000,"locktime":1633824000}]},{"ethAddr":"0xb3d82b1367d362de99ab59a658165aff520cbd4d","avaxAddr":"X-custom1ur873jhz9qnaqv5qthk5sn3e8nj3e0kmzpjrhp","initialAmount":10000000000000000,"unlockSchedule":[{"amount":10000000000000000,"locktime":1633824000}]}],"startTime":1630987200,"initialStakeDuration":31536000,"initialStakeDurationOffset":5400,"initialStakedFunds":["X-custom1g65uqn6t77p656w64023nh8nd9updzmxwd59gh"],"initialStakers":[{"nodeID":"NodeID-4XsLhvvKKgXyBqJbUS9V74eiGDbZf5HYy","rewardAddress":"X-custom18jma8ppw3nhx5r4ap8clazz0dps7rv5u9xde7p","delegationFee":5000},{"nodeID":"NodeID-JqUCBkg87FYDvkiZNSG3hFt3pdsYbLtm4","rewardAddress":"X-custom18jma8ppw3nhx5r4ap8clazz0dps7rv5u9xde7p","delegationFee":5000},{"nodeID":"NodeID-CvwPtxUScTomPZ6o8qhbqSHDRpaMhBD9w","rewardAddress":"X-custom18jma8ppw3nhx5r4ap8clazz0dps7rv5u9xde7p","delegationFee":5000},{"nodeID":"NodeID-3mxMtWgHrWqHcPwEMEJwcS24kzLQbxort","rewardAddress":"X-custom18jma8ppw3nhx5r4ap8clazz0dps7rv5u9xde7p","delegationFee":5000},{"nodeID":"NodeID-CYXgGSJ3VtU6NSRoroJoavVDW3S2DyccV","rewardAddress":"X-custom18jma8ppw3nhx5r4ap8clazz0dps7rv5u9xde7p","delegationFee":5000}],"cChainGenesis":"{\"config\":{\"chainId\":43112,\"homesteadBlock\":0,\"daoForkBlock\":0,\"daoForkSupport\":true,\"eip150Block\":0,\"eip150Hash\":\"0x2086799aeebeae135c246c65021c82b4e15a2c451340993aacfd2751886514f0\",\"eip155Block\":0,\"eip158Block\":0,\"byzantiumBlock\":0,\"constantinopleBlock\":0,\"petersburgBlock\":0,\"istanbulBlock\":0,\"muirGlacierBlock\":0,\"apricotPhase1BlockTimestamp\":0,\"apricotPhase2BlockTimestamp\":0},\"nonce\":\"0x0\",\"timestamp\":\"0x0\",\"extraData\":\"0x00\",\"gasLimit\":\"0x5f5e100\",\"difficulty\":\"0x0\",\"mixHash\":\"0x0000000000000000000000000000000000000000000000000000000000000000\",\"coinbase\":\"0x0000000000000000000000000000000000000000\",\"alloc\":{\"8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC\":{\"balance\":\"0x295BE96E64066972000000\"}},\"number\":\"0x0\",\"gasUsed\":\"0x0\",\"parentHash\":\"0x0000000000000000000000000000000000000000000000000000000000000000\"}","message":"Make time for fun"}`,
				Certificates: []chainv1alpha1.Certificate{
					{
						Cert: `LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUVuVENDQW9XZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFBTUNBWERURTVNVEl6TVRBd01EQXcKTUZvWUR6SXhNakV4TVRBek1URXpOVEkxV2pBQU1JSUNJakFOQmdrcWhraUc5dzBCQVFFRkFBT0NBZzhBTUlJQwpDZ0tDQWdFQTVsWWJmeW9MdVNrbjkrT2ZZcytWbkxQNVpWWHFzRVhPejk4V3NxWC91UnpGaDRrRHdPSjVJN1daCmo3b0tGbUZURW5pb1VDUE1DK1hhcGRlU3h4bHBwaFo4a3VrbzcremcvSk1Sd29FNFUvN1lYZlk1SDg4Y2hFeWQKcDJsOVZLK2J1ZWZyTnd4d3I5R01qWCtUL1g0ZGRGM3RXZXhuK2d6WnVlSzg0UHZBNVlFVDJsaG04eDZhUmJVZQo5aStaWHB1MjVUT3kvMDhrWFJJdzZaV1p3N2lEZTI4blRERDVYMFJEOWozL2UzK0RmMk0xbkVhU0I0VjRveU1yCkc2YkFoQVlONklINlBNV1FZNVBKaHRVSEEzYXFsSDRqRzQ2cTRuYkFHVUlyMlVSRW9CQ0YzNDRLRDlka2R6WkUKOFJ6MkNWWlRXSWlnSkh1dkIxRWFHbDFQNkVkTWxDaDgxQkxGekhySzV2UGVZOFlFZU1lbkJFcm04QVRUaHNicwp6emNFUmVuTi9Ga2RDR1kzUTVKYVFTaTdqbjl4ZksvV3gzTEJoWVNzS1dLbXY3dzNIQmFVeTh1OXkybjErTitJCkdqZ2tjMkpDUFczVEVTeXNDVzlnQW1SN3hsSmk2NTB4SmQ1dnc4czFVTDdoWVFEWU5XYjZ5clFMVHV3UFNMVkwKdHpUUjBSOHBtYjFOUnQ0K01RT2J4ZkhleFpTT0dUSjBHNUhySnJqTkhDMXNEYjZLNVZnZTkvbUJVRVV2RzV6UApSTzMvQVZlSTFteDQ3OThUS0xBbkZZTTNhQTV2VSsxSlZNekxVclY1Q1ZoelZCVXdyWWF4Q0tCNkF6d3MzUDFjCmg2Y1c5UFZrV0JkaE9HMGJsOGdVNk9MUHd4S1dIU3dDdE1LYm5KNUFQN0s1KzFFZ1pxOENBd0VBQWFNZ01CNHcKRGdZRFZSMFBBUUgvQkFRREFnU3dNQXdHQTFVZEV3RUIvd1FDTUFBd0RRWUpLb1pJaHZjTkFRRUxCUUFEZ2dJQgpBSVpnOGpzbm95eWZiT2syazNlR3E2VTZad0lIS0FGaG1YWGFKVVpoWlcvRCtiTkRUdUd5RGRyL3VEcklYSHEwClUxQUdpQXNGZEx2aGVjODV0NXB3bytkWCtrVk1MaEFob3dERXlGbEtMWGxSWG0vUUZNdUtncWFUdFpLVnZQWDUKNXNBUFZJUjQvREhUSlpoZng5L1JoVG11ME12bjY5TFp4MG83N2NGTGZZcWpsUmdudkdLVGRuV3hJWGwxcFlqaAo0SWZ1RmczdmRhYzRsTURpK3l0d2phOFMrMmFPeXNTcGZqbW9NQWNrMk4yU05CSmVYOGRtU3BTNnNMMEdQQ1F0Ck5rdFFNSFBTbW1jcFZybzZEV3FLeG82RnJ2TFNCNkxwMEZacTBKT0dnaFhwbEU0dUpnTm5GZGVQV0tNL1FZUFcKN21ZclY5WEJiemNaSk1IakNiVFZSRC9kazh0NEJISE1PSGxnS1dSeEZHYXovamMzclVvYkVGTUIzRjZ2bVRKdQpSK1c5NjFCQWkzeVUvalZTWC9VcUhZcDRrTkJMeTB6TFAzQUtkSzRMTnd3ZTB6TzdUcUN6azhMMjFPeGY4bkN4CmZOK3pCdjVqWjJtN2FudkpvQVR3UkJONlNuSXR0SU5ocDJTRkE1ZkpEcGdaYzZlK1BmQkFpeWd0Nk5nQ0xMMTkKeE5hYlZWNzE2UlhTWFBlMUFuSlhhN3E1c3NyR3VuY1ZKbjBudk5LL2Z2WjZPSk84RXdxT1lMWjNaNjFBdXpoMApYQVJ2VnBITjhZYWFSclRkUkluZjFwNy93dXRIVU9OaFZzbXJXdnlQdG9kUFpiQ3hLZlJzT0hNVkhhVk85aHdlCnA5YlRkQkxrZ1d2UkI3cW82RGNlT0lZTk9iclhaNE9ZUHBDb3dSV3pUeTVjCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K`,
						Key:  `LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlKUXdJQkFEQU5CZ2txaGtpRzl3MEJBUUVGQUFTQ0NTMHdnZ2twQWdFQUFvSUNBUURtVmh0L0tndTVLU2YzCjQ1OWl6NVdjcy9sbFZlcXdSYzdQM3hheXBmKzVITVdIaVFQQTRua2p0Wm1QdWdvV1lWTVNlS2hRSTh3TDVkcWwKMTVMSEdXbW1GbnlTNlNqdjdPRDhreEhDZ1RoVC90aGQ5amtmenh5RVRKMm5hWDFVcjV1NTUrczNESEN2MFl5TgpmNVA5ZmgxMFhlMVo3R2Y2RE5tNTRyemcrOERsZ1JQYVdHYnpIcHBGdFI3Mkw1bGVtN2JsTTdML1R5UmRFakRwCmxabkR1SU43YnlkTU1QbGZSRVAyUGY5N2Y0Ti9ZeldjUnBJSGhYaWpJeXNicHNDRUJnM29nZm84eFpCams4bUcKMVFjRGRxcVVmaU1ianFyaWRzQVpRaXZaUkVTZ0VJWGZqZ29QMTJSM05rVHhIUFlKVmxOWWlLQWtlNjhIVVJvYQpYVS9vUjB5VUtIelVFc1hNZXNybTg5NWp4Z1I0eDZjRVN1YndCTk9HeHV6UE53UkY2YzM4V1IwSVpqZERrbHBCCktMdU9mM0Y4cjliSGNzR0ZoS3dwWXFhL3ZEY2NGcFRMeTczTGFmWDQzNGdhT0NSellrSTliZE1STEt3SmIyQUMKWkh2R1VtTHJuVEVsM20vRHl6VlF2dUZoQU5nMVp2ckt0QXRPN0E5SXRVdTNOTkhSSHltWnZVMUczajR4QTV2Rgo4ZDdGbEk0Wk1uUWJrZXNtdU0wY0xXd052b3JsV0I3MytZRlFSUzhibk05RTdmOEJWNGpXYkhqdjN4TW9zQ2NWCmd6ZG9EbTlUN1VsVXpNdFN0WGtKV0hOVUZUQ3RockVJb0hvRFBDemMvVnlIcHhiMDlXUllGMkU0YlJ1WHlCVG8KNHMvREVwWWRMQUswd3B1Y25rQS9zcm43VVNCbXJ3SURBUUFCQW9JQ0FRREdpL3J2eHFLTVhUbWlxSWMvZVlpUgpwMkdYUkZRazFrZkxUNVlWTUpvYVN2N0tNZ1VXUXlJQThnMElvMmtHbWFZdUNldXNDTzllWWlmelJMdTArK2JoCjBBaFo2cm5xOXRtSlhveTBpUWF4QU1BcFhwRW5KalNDcGpoUGt0TUNLTTJubG81ZXlVNXBmOHdVUEtDb3BnbGwKd1lGVFBrRHlmaGsvN242NXdNa3FDL1c5Qk83WktzdjR3b09KMnNYdGszUTRaalFwZDJMMUJ6VTZaRVpETzgyNgpuTG5YSjNBTitwNUtxRzZOV2dGVDBZVG96THdiMTZXQm1sTVNaczhUL3RRR1UrUU1kcEJjQll4MXVUTnVmTi9WCldGV1M3NHZGNG03OFZ0bk5VdGVFMGpsVDF2QTliNEdlY1IxRWFaNTd0Zm5xR3Z0UDkzMk1aUkNISVdBNzdSbkkKQ2QwZ0p1WkZVVnpKRGNvRWRVM0Q4ekpGMkE2ZDRVV0cwQUQyMk9OK3RqUGJOM0g4MUlMcTFpOE04Szg0Slc4bQpzNnNxc1BWNFA3WVZqZmdxMkI3NXJhdy9zWXY3NmFQTlVGWC9HVlJLcjlkTHhlUnhQemZEWUpHT2VVSDhtd2w0CmYxUVViZTRLbmVYbU5FeWMxUU5vL283R1BQSCttc0tCRVpUaFZsYVl0Nkx2RXQ4YXVrL2V4SkpRajl4U1o2VmcKa3h5VytZSDkrRkw4K0c2OElmdEtEZDB1TmFOK3JUbXAreVVYRHVrN0h2WE54Sm9hZWhtZzVsNi9jM2dCaG1GRgpVNkZEVmVGZlM1RUFRRnZUbnFnMUQ2TXBzZkZzeHptbVpHWUFPMzhlTGNpckZGMnp2bjdnZGtyT2VLWE5OaHplCmNtU0xGcnBSajd1RFVEQTFzWHpEd1FLQ0FRRUE5c3hXN2U1eUhrT0VoNmZtSHNsR24xR1dnSFZPS243NWpqaTIKQjBpbzh4M01SVjRWY2FvTnQybzFiUHRHdExuYkhlSmdlSjdyRnBjdDZ1NXREYzNkdktMbGc5c1QzUU1IQmpkTApsV3VhOGZySjBkRDZuc2E2bGw0UnF2LzJmU2Q5em5nOXUrdVhnbzEzTnd3VFdCeUsydzdsem5hNG95L3BkZjk0CkVTYVR6QVhzTERSaVBVNE1ONGdGc3NyRFo5Wk8vSGV5cUtZSkxjaVJIelkvMXNvNVJBR3RTQlMyMGM2TGFMVUwKd0hQQ205NDliMzZSV254aVVuY2l4VnJlSTJ3dTcyT0s3UUF4aUZURys3VU9XSU9SSzdSelZySitXOG1aRUtoTAplblRROXFqSk5zUWZGR0pWbDdEOEVZTUR2YjVqc0MrVDVBUEtBSDlZV3FhTVlYRUE1d0tDQVFFQTd1eWtOTkZhCkpVVVZ4Ym03L2tPeENkVzl5bG5sV3paTVZ3ZXo3SFIzOWJkRTBTVFJpWEdJd3JyMUdEdElaN3RkVkZEMm9QUEcKOFhoWWgxWlQ0RmRjMnZFMFhJaDZFNDZ4SjZsTm1LN0pBeXE5dEhwWHl2N1dIN3VZaTNDNlp4OWpzSGQwQnBjaApleEFoZTVFa2lucVo2bXB1Ykd3TzdsTnJIcVNzSXlTKzJLa1Rrd2JpNGRHckUzUkprQWJram1KZ2cyWVlnc2RSCjlFcTVGTXFhRkJleWVOUVl4elJhQVR1NzRuY09uZkhlTEN4VFRKY0tWblRtcXRYSnVzT3FMQndmeEd6TWxEUHoKR0Rqc21ZNmkzQWtvZ0RlbXVJazVaUjNuK2xuN0JXUjhJOU1adUsxeTFnVWZYT0JBUVpMb0Q5eXdUQUwrWU02aQp4VzF2aDQwZ09UYUsrUUtDQVFCRmN0YjdlVi92bUR4UkdEUXZjYUJIOE9PVEhtOXlrZXlUMHUyV095SWYxOERGCnZHWDRhRXdYMHZGWnk1UG9BMnpmaWZadnV2aVlrTTVCRC9yc0tZUStNdkMzSEEwSTRuTTFrcFhZWkVGajJwaTAKVEVSYUxiNFAxa1RPZzl6Tzl5LzF5K3hEVjFaNVRHbkJ1Y292djBocndGTjJ0LzNaSGdCcVRndHhlQk9iRkFlVApvT0lNTWt4SnpDTWVYdVNCOGRLa1JPS25ocUdLbXFnTHltNUllUHVJWVpocmNqakg3WUZaWTZqODdSWlVXa09iCmZsaFV5Qys2MlArVjNhNG85YVozZ3VGek05eThhbTdjSWVUNWozeG9lZzBDMXBPc0xKekFEVHZBSitNdHBlMkoKVmNNUkwySzZudmt2ekZoZktwWk8yL1NYODJFQ3B0TXNIelhkcmJqOUFvSUJBSHpHbFJNSWFsMmdjTGhzUVdPTwprbnlpWlpXeDBQZ0xxVjZpSlRMTnVJQllqOVh4dG9SakNKczU3Qm9WaThDd3R4TDduWEY5SGw2cERRTFE4TWp1Cmx3MjRmakg5REZQK1owSmhScWNBVVBZWWNpNDQzblNqRmN4SXVtZklIWEVSa1l4dE5lampNSmNHVzVZZXZNaWQKTXBpYnNNTnF3M2x2a3pmVHBCcE9iR1RXRitUbTZjSXBMNERmY0RPSmhmOWVIUzFDT25iQ0JXamhSVHM0ZTdNVwpsUnhKR0ErZ3BZaVRXNUh2djNCNUNpQmpuYlVZQkV3V2pRaVcwZDE1cGZ1WFRIZldvaGliOE02cm05U3VDeHVDCnBPWWhLaTZoYTVvRlBrc2VodHZRR0l6VkNFL01OWGJVQWdjTkRrR3dxUVR2cWhwb1RkVGMxV0Rwd0I4NGNxV3UKZUhFQ2dnRUJBSTVlTnQrdWdxMjJta0hkeU5rRFdSaVBCc2pYVHJwaisrbkZMV0NINSs5RndIUUp3Z1VWU3Z2TApOWkN5eC9rVDVIbzlJUTU1MTBvQzNrd2J3MkhaK1JNRnMrYloyTnNOZkxvQXN1T1ZkaWdGbmFvSCtNUWUwcnYxCnpvTitSaFFRTzJuQmJobWhPZ1VCMWhUZnoyUXgxaFRIWVQ4alhwenJKKzVhLzZqd0NqSElsOUFkNkxobk1KelgKUGc2TTVXeURpLzN4aHl0WEJxQ3liaTFERVVYQzdSeXpxdkxuSGJWdnBsV21qMzNycHVwa2htQlRLU2cxZnZHZgoxYXIwTEpmdFQ3Q1NqM24rQWFRWkVmbXBDSzhTSDVDa3haME1FWlRSTXU5dk54VFYxRk1DM3FqVE1xNVJqSzdECkNNUmx0MXdkUVZHb214TldoZ1c1N0ZQMWhZUEtoQnM9Ci0tLS0tRU5EIFJTQSBQUklWQVRFIEtFWS0tLS0tCg==`,
					},
				},
				Env: []corev1.EnvVar{
					{
						Name:  "AVAGO_LOG_LEVEL",
						Value: "debug",
					},
				},
			}
			keyBootstrapper := types.NamespacedName{
				Name:      AvalanchegoValidatorName,
				Namespace: AvalanchegoNamespace,
			}
			toCreateBootstrapper := &chainv1alpha1.Avalanchego{
				TypeMeta: metav1.TypeMeta{
					Kind:       AvalanchegoKind,
					APIVersion: AvalanchegoAPIVersion,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      keyBootstrapper.Name,
					Namespace: keyBootstrapper.Namespace,
				},
				Spec: specBootstrapper,
			}

			specWorker := chainv1alpha1.AvalanchegoSpec{
				Tag:            "v1.6.3",
				DeploymentName: AvalanchegoWorkerDeploymentName,
				NodeCount:      1,
				Genesis:        `{"networkID":12346,"allocations":[{"ethAddr":"0xb3d82b1367d362de99ab59a658165aff520cbd4d","avaxAddr":"X-custom1g65uqn6t77p656w64023nh8nd9updzmxwd59gh","initialAmount":0,"unlockSchedule":[{"amount":10000000000000000,"locktime":1633824000}]},{"ethAddr":"0xb3d82b1367d362de99ab59a658165aff520cbd4d","avaxAddr":"X-custom18jma8ppw3nhx5r4ap8clazz0dps7rv5u9xde7p","initialAmount":300000000000000000,"unlockSchedule":[{"amount":20000000000000000},{"amount":10000000000000000,"locktime":1633824000}]},{"ethAddr":"0xb3d82b1367d362de99ab59a658165aff520cbd4d","avaxAddr":"X-custom1ur873jhz9qnaqv5qthk5sn3e8nj3e0kmzpjrhp","initialAmount":10000000000000000,"unlockSchedule":[{"amount":10000000000000000,"locktime":1633824000}]}],"startTime":1630987200,"initialStakeDuration":31536000,"initialStakeDurationOffset":5400,"initialStakedFunds":["X-custom1g65uqn6t77p656w64023nh8nd9updzmxwd59gh"],"initialStakers":[{"nodeID":"NodeID-4XsLhvvKKgXyBqJbUS9V74eiGDbZf5HYy","rewardAddress":"X-custom18jma8ppw3nhx5r4ap8clazz0dps7rv5u9xde7p","delegationFee":5000},{"nodeID":"NodeID-JqUCBkg87FYDvkiZNSG3hFt3pdsYbLtm4","rewardAddress":"X-custom18jma8ppw3nhx5r4ap8clazz0dps7rv5u9xde7p","delegationFee":5000},{"nodeID":"NodeID-CvwPtxUScTomPZ6o8qhbqSHDRpaMhBD9w","rewardAddress":"X-custom18jma8ppw3nhx5r4ap8clazz0dps7rv5u9xde7p","delegationFee":5000},{"nodeID":"NodeID-3mxMtWgHrWqHcPwEMEJwcS24kzLQbxort","rewardAddress":"X-custom18jma8ppw3nhx5r4ap8clazz0dps7rv5u9xde7p","delegationFee":5000},{"nodeID":"NodeID-CYXgGSJ3VtU6NSRoroJoavVDW3S2DyccV","rewardAddress":"X-custom18jma8ppw3nhx5r4ap8clazz0dps7rv5u9xde7p","delegationFee":5000}],"cChainGenesis":"{\"config\":{\"chainId\":43112,\"homesteadBlock\":0,\"daoForkBlock\":0,\"daoForkSupport\":true,\"eip150Block\":0,\"eip150Hash\":\"0x2086799aeebeae135c246c65021c82b4e15a2c451340993aacfd2751886514f0\",\"eip155Block\":0,\"eip158Block\":0,\"byzantiumBlock\":0,\"constantinopleBlock\":0,\"petersburgBlock\":0,\"istanbulBlock\":0,\"muirGlacierBlock\":0,\"apricotPhase1BlockTimestamp\":0,\"apricotPhase2BlockTimestamp\":0},\"nonce\":\"0x0\",\"timestamp\":\"0x0\",\"extraData\":\"0x00\",\"gasLimit\":\"0x5f5e100\",\"difficulty\":\"0x0\",\"mixHash\":\"0x0000000000000000000000000000000000000000000000000000000000000000\",\"coinbase\":\"0x0000000000000000000000000000000000000000\",\"alloc\":{\"8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC\":{\"balance\":\"0x295BE96E64066972000000\"}},\"number\":\"0x0\",\"gasUsed\":\"0x0\",\"parentHash\":\"0x0000000000000000000000000000000000000000000000000000000000000000\"}","message":"Make time for fun"}`,
				Certificates: []chainv1alpha1.Certificate{
					{
						Cert: `LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUVuVENDQW9XZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFBTUNBWERURTVNVEl6TVRBd01EQXcKTUZvWUR6SXhNakV4TVRBek1URXpOVEkyV2pBQU1JSUNJakFOQmdrcWhraUc5dzBCQVFFRkFBT0NBZzhBTUlJQwpDZ0tDQWdFQXkzOWR5Sk5pQVVUT3p5UVhPR3J3Yk5FaFVqZDRuUmJDNS8xVWU4YW5sMmVCVVBubEZwWkhJY2xZCkFWL2pvNGVSZGhING5oMDV6UU9ydHdzcHkzdGNjWjFJVDBEa1N1c1NjRG1xdHVZdFUxeStnS3R0TjMvanUvbG8KbG11bGZabFEvL29JekNWRExqR1hkOFhpQlhVd2dKU1E5MFhvMTVISU1aMUUxcVhVT3ZzN0ZPb0xkeHZ2WHpGVQpsUHEzVVI5dGJUNjJBRS9KenZlZGdGSEtQMWs2a0tRQjFzNXB5ZkZEOFR4czVsbE9YUGxqZzdHNUtYaTdWMURCCmpFam91amJmMWdobUNYY1hhNGtSWkNvditrdktqSHF0RVUwQ1E3OVpRdGVzWGN5UFhSOXBVQ1ArWkJuMEFwTTYKWVNWWW9ZaXIvWWJSZjdlNGNyc1dEUDdUbFIvdmdIQ21zazZaeExod1dKZUZmZTNMV2cvNzN6OUlXT2tyR3NNQQpsNWdBMnA2aTZtbDE1RXpvRUNXM1NrZm1kdG9heWJneTFOMjVvZ1BEUVBIbUwxVEJIV1l1VUk3aVU0YjBuWFVnCkNyRzBjQk05U1BrQ1BpWVdha3hRbkNXekFwYkZNeU8wVzdKN3EwVGZFc25Mby93aTBsK2VzUTQvcFptWHJkNmEKWDNFNVVLOGtwTUxodmNXc3BvOXpqOXRVMEwzVFhEcFBkMUxnQ1RhMUM5aWhFZllwd29RcFl5N1hrMlh6eFpuNQpDUkZGM0F4TFdOaUNjRk85Y2JRVGx6TVVzaVhlajJmTWlnRHpDREIyaGlucDZNRklsbGlTZ0VVcGVNOFIzcVZGCnFxN0gzTmNKNVhlWlJwckhHNkpOeEdBYmVkam1WWXVTYkduMnlBKzJhODNIaDJOTU12TUNBd0VBQWFNZ01CNHcKRGdZRFZSMFBBUUgvQkFRREFnU3dNQXdHQTFVZEV3RUIvd1FDTUFBd0RRWUpLb1pJaHZjTkFRRUxCUUFEZ2dJQgpBR2s3RGM0MmNYT3FTV0FhNTgxbDRUbzdMdE9oeTE1MmJKM3cxcHNrakFUMStEQ204VEphZTdYdm92YzdTU2FqCk5wZ1BNVFNFaEZjRWhwNzBpWHA3V3k4QVd2MTlSSldFakMwSFRCS3BOUy9CS3V3NWtLZWhlbW9scHUrVjdsdUwKMTduSzl3bGRpTWxEWTJzSk9zR3RaWHF2OHExZkFBS0lySys0VmZBSDJKVmZQUUJhTStDVmc2ZTk1TVUvSFRTSgpGbG96a2NzYi8vdnZvTk9HR0RjQlFtSFB3QStZUnhLTThYTmNlM1l4cDkyT2I0dGYyY25yemtlall4OUNZYnJuCk5VOVVsYnZteDljVlFsK1g0NHBKOGtaNFlLNndNQmowb3doTDlIc0o3YVlseUQyWVdGMFVhV0R4NVBUTDVoMnYKTDdka2xqVXQwWlVIZ1J4NCtQc2JFb2pyYUlwV0tYMUVKUzRDVUwvTzNMZjVWZFhZdURaK0lTY0hlTFJ4WHc4UgpZNjZPWnJJdEpjbWFkVHFaR2hoYUN2K1psM2c1UjU4NFZVZlhYTXpSNVBDeGJWQ1lpUzQyWHc0YmZFVjk0ZlhoCnR1TXV3SGJhQ1dmK2FlcFJISWVLS1RkM1YzQmFGbUZMYktLT1lyZ2EreDFMMFJraDdwNEcvajhWSkRrRHVoOWwKeHU0Vk5ZaEZERThCcnViczJnaXRhbGZpY2EwTjB6cFR3cWhKOGpPMmtxVUVCdGdETlYvR3Y3UjF6TC80NU9oTApkWUtqU1REbDNrL3lRNE91OW1KTit5UHI5WUhSTFkwTzk1SytEdjAzUW4zcTAveC9LYTRjcVhSZlVMek1Zc2UwClBmWDNjU0RWalc0SEVvRmRIeXZXSDlTK1IrSnRHSHQ0R2VUS1RtV3E0VUZpCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K`,
						Key:  `LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlKUWdJQkFEQU5CZ2txaGtpRzl3MEJBUUVGQUFTQ0NTd3dnZ2tvQWdFQUFvSUNBUURMZjEzSWsySUJSTTdQCkpCYzRhdkJzMFNGU04zaWRGc0xuL1ZSN3hxZVhaNEZRK2VVV2xrY2h5VmdCWCtPamg1RjJFZmllSFRuTkE2dTMKQ3luTGUxeHhuVWhQUU9SSzZ4SndPYXEyNWkxVFhMNkFxMjAzZitPNytXaVdhNlY5bVZELytnak1KVU11TVpkMwp4ZUlGZFRDQWxKRDNSZWpYa2NneG5VVFdwZFE2K3pzVTZndDNHKzlmTVZTVStyZFJIMjF0UHJZQVQ4bk85NTJBClVjby9XVHFRcEFIV3ptbko4VVB4UEd6bVdVNWMrV09Ec2JrcGVMdFhVTUdNU09pNk50L1dDR1lKZHhkcmlSRmsKS2kvNlM4cU1lcTBSVFFKRHYxbEMxNnhkekk5ZEgybFFJLzVrR2ZRQ2t6cGhKVmloaUt2OWh0Ri90N2h5dXhZTQovdE9WSCsrQWNLYXlUcG5FdUhCWWw0Vjk3Y3RhRC92ZlAwaFk2U3Nhd3dDWG1BRGFucUxxYVhYa1RPZ1FKYmRLClIrWjIyaHJKdURMVTNibWlBOE5BOGVZdlZNRWRaaTVRanVKVGh2U2RkU0FLc2JSd0V6MUkrUUkrSmhacVRGQ2MKSmJNQ2xzVXpJN1Jic251clJOOFN5Y3VqL0NMU1g1NnhEaitsbVpldDNwcGZjVGxRcnlTa3d1Rzl4YXltajNPUAoyMVRRdmROY09rOTNVdUFKTnJVTDJLRVI5aW5DaENsakx0ZVRaZlBGbWZrSkVVWGNERXRZMklKd1U3MXh0Qk9YCk14U3lKZDZQWjh5S0FQTUlNSGFHS2Vub3dVaVdXSktBUlNsNHp4SGVwVVdxcnNmYzF3bmxkNWxHbXNjYm9rM0UKWUJ0NTJPWlZpNUpzYWZiSUQ3WnJ6Y2VIWTB3eTh3SURBUUFCQW9JQ0FITk1Pc3JHRnFVNVl5T2lBellJQVNqbQpaTWE4Znk0aUUxUjJDRVFKRGpPT2hZcG56QkM4SEpsY0J1emdjNDNYNWViTHo5MW1HYlc2K3JPL00zTUM5aUc1ClI1ci8zVmxGVHpFZXUwYmRxNWlyMTVQM2pPNEJHL3NKR09VQklNYkU4MHZWVXQ2M3poU0NMSnZFRm9lWkdsMy8KenhNTEhSM21qMUx0RkcrNWpVSE56bS9QRzZma3YvOWpaOVR4S0tSaDloSUxrZnNqT2VoMkMxc0UvRjVnSS9xSApzak1PeUltT2xUdzlURVpIRzBzNlVkUHdBa1VwRHB3dU9UdE9vKzI5NFp6WExWajNqT0YweTlIQXhXWS9Rd2ZOCkNmbmZkQVVHaVlDQnlqdHJCMTl2eUsrTGRUc3FLVUs4UUR1Q2VYRXNpcVllbU55UUw0VngwdENTSVRkQTNPVVoKc0syb3pPYWxoWER5OExsTmZKNnY4YXdCbFg3cXZ4UDhJSUpYcVJlM3RqWFlOajRVU1dPWEYreHVINnAzdStYYwowVDhRZ0xYZ2dvMTU5eEM4RkNQWXkzOTJpSCtZZWtPNUIxaGpVRENRMDJjNGVwejRrQmhnemdDak5mc0F4MmYzCld1OUFTaEVLRmRBSmVMQ0NXWjhqS0l4a0VseEFhd1MyT0dGN1lQdjdITEFvNU9BeFZ5MHVsK3JQUDB6Unl5SFMKcnp1ME9rTUZWbFJuUUdKUE12OTlsYjdaWHBTajlrWkloQWt6QUlSclpHV1U1d052Nkt3M21SV3BkSWI3NEd1TgpqczJvMW5kS2Z6R3pUMi96TmZPczNmUU1LRkJEMzJ2V0NHMzhWYUhCTzltbVhpU0pqNXVCV0piMEQyNXhXYmhOCjBMdVNaRzRmSXZUb1p2cCtSYnloQW9JQkFRRFZOaDRCSWROM2NnY0Rmb0IxcDg5UHRmZVZtVklZYUpxY3RnV0EKdkxFTG5QQS9oeG1ESHU2MFgxSE9PWUlQc2E1akdFS0xLcWJreEMyWVliWUc1bWpFYUNUR1N2eVZDaG4xTHdWSAo3Rjk5N01lb3lWYm1adW0ydGFadWIvZVAwZGxuZy92QlNUOHpxRVltdndrcWRBVEplN2dIMUt3TmRiL041c0RNCnNPTEFrbTVFR3RxWllEY2p5cG9iMGJteG5sS2o5K2daNkRMd2RmUzFacVd4ckt4RVFTNTFVOXJVdGR4UFZYK2kKdWIrdWhFdnVTZTdmQndQemlzOCtzdXRxRUh0bmdEVllEY2lnS0ttcEU3VHNGUVdicjBzQmdwY01SdmJndjZCNQpWSlRXUi9CazBXRG9DNWFCek9VajdRU1g5MlJzSnV1d2pBQXlSMmpncXVJWW1yYkhBb0lCQVFEMFZqRThHdEZVCmk0MXlKcDZkb1hWdUVBdjh2SFFNSmNLUHFTVE5DRGF5RCtUd3JvVnhCSjJldGkyTUp0Z0JEVzNJbW9mdy9kcFMKd1lsMXE0OVJFeWdsa2d4VTJYdWR0Yk9uazVSYkJnOHlpR2RyU21DWU00a3VxNkpYWXBxVFZIajlIUTE5emZMcgpUWjVsRE93S2N5U2tnYTBneWxXNEdJU3ZFZ3ZneGJYb2NjSnBWNXNEL0NQdDVtUjZCOE9RVFlCTGhINTFrY2RVCmpMeHN6SXRTaTBWSzM5SElqOWRUZWVkc3RWY0c5SVc5WWVQbUJsNHFGbk8wK0c3T3F5OGNZb0RIcW00SWlGMzUKaks1WlpMTU9ZalROcEs4ZEg3a0FHVjFmOG9kZ2VUazZ0c3ZVZkRTRUlEbVdFSk5CVW0zYmg5R3BCZ0pDeDVNago0dDBEQ1dCelFnWjFBb0lCQUJUVENWRXcvWmVBQXFGYnZLNUJLcVZ0YjNZa0dIbWIxZVlTZlMwYXdPd1Njd0N4CmNGTjNOUGRYREFWcFpvT2o1aFYxckNJdGswbHF3ODFMVmQwTXFoVHMyeEtuQms2RVF3N0lmZXFOY3JJNDZ6TlkKSHUyNEJZRzc4anA5SXgvZjdpMEhIaEs5MWJkMDZ3MGp3WUJzL242elg4RWNDNFh4QnovVUZ1YW5MQzZFM3RJMgpFVDNEd1A3MDdlSmp0SkJkbDFLK2h1UG80dmpMZkpBdksyWFVLS3N0OXB5dENRV1hrYUlLQnNKZEJCVEdoU2dMCi9wRzMvTEhQei9nZXY1R0hkSllpVnBONEhTMVBhMnJCS3YyWC9BazlzTVMvL1lMTWQ1WnlBUGw0d21TL2VBSlEKMVBjMUVva3crdnhzVFBPT3pUY25BZ1FuV0dtUXdmU1huQ2V0RE1jQ2dnRUJBTnYyS0ZhNnNjNlIzMkZ2WVFYNQpUNVVvL3hHa3VqZ2hXamtvaFlmTEtDbys0dFRGMkQyNWNRaHJheStyM0hOK0dtSW9zODhCU1NXTk0rbHA3QmlKCnpXK2RQbHE0ZTIrc0h6THlTZkZ6MEFTbkJhdHlCdW1lSTVhUFR4T3FJZ3dXVk9GUTRVOXJNNUFmalVQZFVUWEwKR0thOFV4YWM4SFJPSmt6UlN6NHIzeXFHRndYc3B4SDhVSUFnRkQ1RGRRd1lxVEhTOG1GM1BtSmdYRlQ2QTBicApPQlZDejBIbU5Hdmk1N05Xd1NUeXh0K0tHN2Q5N2hHbnFyeTFsbE9aaWt6Y1pLRGJUam1DUWsraEZXaEdubWVKCmc0M0oveGVSOG1NamNvc285RFNtalIzTmFFdy9FS3dOc3Fuay9Cd25UOXo5TllNYmRMZVhvV0FDSVFOVjBxMlEKTiswQ2dnRUFRS2xPR1dzMWV1V3AyRTQ0RXN5RXZDQjd5cGJEdUF3TG9saGZnTVlIWmFFd2RVdU80R1BrSjVZTgpwTWVwUVRQa0VWc1NOb2poVTgxTHFNbGpWKzVwc1NVTnJTS2owUkc1aE5GcWViQmlyNUVXTVl5MlhtTFFDSHhvClNnOTBXWUZseEk0S0JXZmhiZTBTTTY0VUdHRjNCMzBiL052MWI4SW13MTNFeXRwcUZ5Wm1naG9wUk1rcXZFSWkKZzNCcDJvQXRkb2ZoNXUrQ1hTSlZJTVJCUkhwWlpDdUdBcDZTVklkTVVaYWdSSzZtbitVYXVHVzRBV1dIaW1sQQpHRjBTdlJKbGJLaXFuUE42eG5XZzZ2SzZ2T0Vpa000T25sTURTaHlLVFpTN2poSElUWXczTk9IczR2ek1mZ1JHCmdHYjdvMTFTeHk4Q0xWNlgyOExzbnUvaWN2ZTlKQT09Ci0tLS0tRU5EIFJTQSBQUklWQVRFIEtFWS0tLS0tCg==`,
					},
				},
				Env: []corev1.EnvVar{
					{
						Name:  "AVAGO_LOG_LEVEL",
						Value: "debug",
					},
				},
			}
			keyWorker := types.NamespacedName{
				Name:      AvalanchegoWorkerName,
				Namespace: AvalanchegoNamespace,
			}
			toCreateWorker := &chainv1alpha1.Avalanchego{
				TypeMeta: metav1.TypeMeta{
					Kind:       AvalanchegoKind,
					APIVersion: AvalanchegoAPIVersion,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      keyWorker.Name,
					Namespace: keyWorker.Namespace,
				},
				Spec: specWorker,
			}

			By("Creating Avalanchego Bootstrapper successfully")
			Expect(k8sClient.Create(context.Background(), toCreateBootstrapper)).Should(Succeed())
			time.Sleep(time.Second * 5)

			Eventually(func() bool {
				f := &chainv1alpha1.Avalanchego{}
				_ = k8sClient.Get(context.Background(), keyBootstrapper, f)
				return f.Status.Error == ""
			}, timeout, interval).Should(BeTrue())

			By("Checking if Bootstrapper's Uri is exposed")

			fetchedBootstrapper := &chainv1alpha1.Avalanchego{}

			Eventually(func() bool {
				_ = k8sClient.Get(context.Background(), keyBootstrapper, fetchedBootstrapper)
				return fetchedBootstrapper.Spec.NodeCount == len(fetchedBootstrapper.Status.NetworkMembersURI)
			}, timeout, interval).Should(BeTrue())

			By("Creating Avalanchego Worker successfully")
			toCreateWorker.Spec.BootstrapperURL = fetchedBootstrapper.Status.NetworkMembersURI[0]
			Expect(k8sClient.Create(context.Background(), toCreateWorker)).Should(Succeed())
			time.Sleep(time.Second * 5)

			Eventually(func() bool {
				f := &chainv1alpha1.Avalanchego{}
				_ = k8sClient.Get(context.Background(), keyWorker, f)
				return f.Status.Error == ""
			}, timeout, interval).Should(BeTrue())

			By("Checking if Worker's Uri is exposed")

			fetchedWorker := &chainv1alpha1.Avalanchego{}

			Eventually(func() bool {
				_ = k8sClient.Get(context.Background(), keyWorker, fetchedWorker)
				return fetchedWorker.Spec.NodeCount == len(fetchedWorker.Status.NetworkMembersURI)
			}, timeout, interval).Should(BeTrue())

			By("Deleting the scope")
			Eventually(func() error {
				f := &chainv1alpha1.Avalanchego{}
				_ = k8sClient.Get(context.Background(), keyBootstrapper, f)
				return k8sClient.Delete(context.Background(), f)
			}, timeout, interval).Should(Succeed())

			Eventually(func() error {
				f := &chainv1alpha1.Avalanchego{}
				return k8sClient.Get(context.Background(), keyBootstrapper, f)
			}, timeout, interval).ShouldNot(Succeed())

			Eventually(func() error {
				f := &chainv1alpha1.Avalanchego{}
				_ = k8sClient.Get(context.Background(), keyWorker, f)
				return k8sClient.Delete(context.Background(), f)
			}, timeout, interval).Should(Succeed())

			Eventually(func() error {
				f := &chainv1alpha1.Avalanchego{}
				return k8sClient.Get(context.Background(), keyWorker, f)
			}, timeout, interval).ShouldNot(Succeed())

		})
	})

})
