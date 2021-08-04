package prometheus

/*
var (
	SampleResultStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "hidra",
			Name:      "sample_metric_status",
			Help:      "This is my sample metric status",
		},
		[]string{"agent_id", "sample_id", "sample_name", "checksum"},
	)

	SampleResultTime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "hidra",
			Name:      "sample_metric_time",
			Help:      "This is my sample metric status",
		},
		[]string{"agent_id", "sample_id", "sample_name", "checksum"},
	)

	sampleStepResultTime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "hidra",
			Name:      "sample_step_metric_time",
			Help:      "This is my sample metric status",
		},
		[]string{"agent_id", "sample_id", "sample_name", "checksum", "step_name", "params", "negate"},
	)
)

func StartPrometheus(listenAddr string, pullTime int) {
	prometheus.MustRegister(SampleResultStatus)
	prometheus.MustRegister(SampleResultTime)
	prometheus.MustRegister(sampleStepResultTime)

	go func() {
		for {
			samples, _ := models.GetSamples()
			agents, _ := models.GetAgents()

			for _, sample := range samples {
				sampleDef, _ := models.ReadScenariosYAML(sample.SampleData)
				for _, agent := range agents {
					lastMetricByAgent, _ := sample.GetLastMetricByAgent(agent.ID.String())

					sampleStatus := 1
					if len(lastMetricByAgent.Error) == 0 {
						sampleStatus = 0

					}

					// Write test status
					SampleResultStatus.WithLabelValues(
						agent.ID.String(),
						sample.ID.String(),
						sample.Name,
						sample.Checksum).Set(float64(sampleStatus))

					testTime := float64(lastMetricByAgent.EndDate.UnixNano() - lastMetricByAgent.StartDate.UnixNano())
					SampleResultTime.WithLabelValues(
						agent.ID.String(),
						sample.ID.String(),
						sample.Name,
						sample.Checksum).Set(testTime)

					StepResults, _ := lastMetricByAgent.GetStepResults()

					for i := 0; i < len(StepResults); i++ {
						StepResult := StepResults[i]
						step := sampleDef.Scenario.Steps[i]

						paramsStr, _ := json.Marshal(step.Params)

						StepResultTime := float64(StepResult.EndDate.UnixNano() - StepResult.StartDate.UnixNano())
						sampleStepResultTime.WithLabelValues(
							agent.ID.String(),
							sample.ID.String(),
							sample.Name,
							sample.Checksum,
							step.Type,
							string(paramsStr),
							strconv.FormatBool(step.Negate)).Set(StepResultTime)
					}
				}
			}

			time.Sleep(time.Second * time.Duration(pullTime))
		}
	}()

	// Run HTTP metrics server with auth basic
	log.Println("Starting metrics at " + listenAddr)
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}
*/
