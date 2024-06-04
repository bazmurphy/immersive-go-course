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

	cronJobsFailedExecution = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "cron_jobs_execution_failed_total",
		Help: "total number of cron jobs failed to execute",
	})

	cronJobExecutionDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "cron_job_execution_time_seconds",
		Help: "execution time of each cron job in seconds",
	}, []string{"topic"})

	cronJobExecutionDelayStart = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "cron_job_execution_delay_start_seconds",
		Help: "the delay in seconds, from when the record was first produced, to when it started execution on the consumer",
	}, []string{"topic"})

	cronJobExecutionDelayEnd = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "cron_job_execution_delay_end_seconds",
		Help: "the delay in seconds, from when the record was first produced, to when it finished execution on the consumer",
	}, []string{"topic"})

	cronJobsRetried = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "cron_jobs_retried_total",
		Help: "total number of cron jobs retried",
	})

	cronJobsRetryFailed = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "cron_jobs_retry_failed_total",
		Help: "total number of cron jobs retried that ultimately failed to execute",
	})
)

func init() {
	prometheus.MustRegister(cronJobsConsumed)
	prometheus.MustRegister(cronJobsExecuted)
	prometheus.MustRegister(cronJobsFailedExecution)
	prometheus.MustRegister(cronJobExecutionDuration)
	prometheus.MustRegister(cronJobExecutionDelayStart)
	prometheus.MustRegister(cronJobExecutionDelayEnd)
	prometheus.MustRegister(cronJobsRetried)
	prometheus.MustRegister(cronJobsRetryFailed)
}
