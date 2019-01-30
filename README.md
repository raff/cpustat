# cpustat
A tool to chart CPU usage

Usage:

    cpustat [-all] [-wait=5s]

    Usage of cpustat:
      -all
          display all stat categories (user, nice, system, idle, iowait, irq, steal)
      -wait duration
          wait between runs (default 5s)

- By default it displays work time (user + system, etc.), steal time and idle time. Use -all to display more details.
- Wait is expressed as "Go" time.Duration (i.e. use 1.5s or 1500ms for 1.5 seconds).
