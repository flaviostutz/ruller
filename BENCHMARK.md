# Benchmark

  * 4 cores, 8GB ram, Docker
  * benchmark command

```
ab -p sample-rules/ab-post.txt -T application/json -c 6 -n 10000 http://localhost:3000/test
```

### No data returned by evaluated rules

#### Test with a group of 100 rules (each request evaluates all rules)
```
Concurrency Level:      6
Time taken for tests:   6.190 seconds
Complete requests:      10000
Failed requests:        0
Total transferred:      1090000 bytes
Total body sent:        1570000
HTML transferred:       20000 bytes
Requests per second:    1615.64 [#/sec] (mean)
Time per request:       3.714 [ms] (mean)
Time per request:       0.619 [ms] (mean, across all concurrent requests)
Transfer rate:          171.98 [Kbytes/sec] received
                        247.71 kb/s sent
                        419.69 kb/s total
```

#### Test with a group of 1000 rules (each request evaluates all rules)
```
Concurrency Level:      6
Time taken for tests:   7.631 seconds
Complete requests:      10000
Failed requests:        0
Total transferred:      1090000 bytes
Total body sent:        1570000
HTML transferred:       20000 bytes
Requests per second:    1310.44 [#/sec] (mean)
Time per request:       4.579 [ms] (mean)
Time per request:       0.763 [ms] (mean, across all concurrent requests)
Transfer rate:          139.49 [Kbytes/sec] received
                        200.92 kb/s sent
                        340.41 kb/s total
```

#### Test with a group of 5000 rules (each request evaluates all rules)
```
Concurrency Level:      6
Time taken for tests:   17.790 seconds
Complete requests:      10000
Failed requests:        0
Total transferred:      1090000 bytes
Total body sent:        1570000
HTML transferred:       20000 bytes
Requests per second:    562.11 [#/sec] (mean)
Time per request:       10.674 [ms] (mean)
Time per request:       1.779 [ms] (mean, across all concurrent requests)
Transfer rate:          59.83 [Kbytes/sec] received
                        86.18 kb/s sent
                        146.02 kb/s total
```

#### Test with a group of 10000 rules (each request evaluates all rules)
```
Concurrency Level:      6
Time taken for tests:   35.488 seconds
Complete requests:      10000
Failed requests:        0
Total transferred:      1090000 bytes
Total body sent:        1570000
HTML transferred:       20000 bytes
Requests per second:    281.78 [#/sec] (mean)
Time per request:       21.293 [ms] (mean)
Time per request:       3.549 [ms] (mean, across all concurrent requests)
Transfer rate:          29.99 [Kbytes/sec] received
                        43.20 kb/s sent
                        73.20 kb/s total
```

### Some data returned by each evaluated rule

#### Test with a group of 100 rules (each request evaluates all rules)
```
Concurrency Level:      6
Time taken for tests:   12.179 seconds
Complete requests:      10000
Failed requests:        0
Total transferred:      152320000 bytes
Total body sent:        1560000
HTML transferred:       151440000 bytes
Requests per second:    821.11 [#/sec] (mean)
Time per request:       7.307 [ms] (mean)
Time per request:       1.218 [ms] (mean, across all concurrent requests)
Transfer rate:          12213.96 [Kbytes/sec] received
                        125.09 kb/s sent
                        12339.05 kb/s total
```

### Test with a group of 1000 rules (each request evaluates all rules)
```
Concurrency Level:      6
Time taken for tests:   56.678 seconds
Complete requests:      10000
Failed requests:        0
Total transferred:      1526310000 bytes
Total body sent:        1560000
HTML transferred:       1525430000 bytes
Requests per second:    176.43 [#/sec] (mean)
Time per request:       34.007 [ms] (mean)
Time per request:       5.668 [ms] (mean, across all concurrent requests)
Transfer rate:          26298.18 [Kbytes/sec] received
                        26.88 kb/s sent
                        26325.06 kb/s total
```

### Test with a group of 5000 rules (each request evaluates all rules)
```
Concurrency Level:      6
Time taken for tests:   170.733 seconds
Complete requests:      10000
Failed requests:        0
Total transferred:      7671840000 bytes
Total body sent:        1560000
HTML transferred:       7670960000 bytes
Requests per second:    58.57 [#/sec] (mean)
Time per request:       102.440 [ms] (mean)
Time per request:       17.073 [ms] (mean, across all concurrent requests)
Transfer rate:          43881.65 [Kbytes/sec] received
                        8.92 kb/s sent
                        43890.58 kb/s total
```

### Test with a group of 10000 rules (each request evaluates all rules)
```
Concurrency Level:      6
Time taken for tests:   302.636 seconds
Complete requests:      10000
Failed requests:        0
Total transferred:      15353300000 bytes
Total body sent:        1560000
HTML transferred:       15352420000 bytes
Requests per second:    33.04 [#/sec] (mean)
Time per request:       181.581 [ms] (mean)
Time per request:       30.264 [ms] (mean, across all concurrent requests)
Transfer rate:          49542.93 [Kbytes/sec] received
                        5.03 kb/s sent
                        49547.96 kb/s total
 ```
