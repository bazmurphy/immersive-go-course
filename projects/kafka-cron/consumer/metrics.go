package main

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	cronJobsConsumed = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "cron_jobs_consumed_total",
		Help: "total number of cron jobs consumed successfully",
	})

	cronJobsExecuted = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "cron_jobs_executed_total",
		Help: "total number of cron jobs executed successfully",
	})

	cronJobExecutionTime = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "cron_job_execution_time_seconds",
		Help: "execution time of each cron job command in seconds",
	}, []string{"cron_job_id", "cron_job_command", "topic"})

	cronJobsFailedExecution = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "cron_jobs_execution_failed_total",
		Help: "total number of cron jobs failed to execute",
	})

	cronJobsRetried = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "cron_jobs_retried_total",
		Help: "total number of cron jobs retried",
	})
)

func init() {
	prometheus.MustRegister(cronJobsConsumed)
	prometheus.MustRegister(cronJobsExecuted)
	prometheus.MustRegister(cronJobExecutionTime)
	prometheus.MustRegister(cronJobsFailedExecution)
	prometheus.MustRegister(cronJobsRetried)
}
