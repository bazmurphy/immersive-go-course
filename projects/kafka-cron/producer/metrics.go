package main

import "github.com/prometheus/client_golang/prometheus"

var (
	cronJobsScheduled = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "cron_jobs_scheduled_total",
		Help: "total number of cron jobs scheduled successfully",
	})

	cronJobSchedulingFailures = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "cron_job_schedule_failures_total",
		Help: "total number of failures when scheduling cron jobs",
	})

	cronJobsProduced = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "cron_jobs_produced_total",
		Help: "total number of cron jobs produced successfully",
	})

	cronJobProducingFailures = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "cron_jobs_producing_failures_total",
		Help: "total number of failures when producing cron jobs",
	})
)

func init() {
	prometheus.MustRegister(cronJobsScheduled)
	prometheus.MustRegister(cronJobSchedulingFailures)
	prometheus.MustRegister(cronJobsProduced)
	prometheus.MustRegister(cronJobProducingFailures)
}
