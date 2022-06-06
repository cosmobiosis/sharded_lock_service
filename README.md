# Sharded Lock Service
Sharded lock service consists of a group of lock servers that runs on grpc protocol.
It's a separate service that communicates with our Tribbler 2.0 which supports transactions.
To run the whole system, please use the following instructions.

## Instructions
### Tribbler 2.0 Download
Please first download our Tribbler 2.0 and use the git branch final_proj
> git clone https://github.com/ucsd-cse223b-sp22/lab-3-trinityforce.git

### Run golang lock servers
>go mod download

>go run cmd/LockServerExec/main.go

### Run Tribber 2.0 chaotic test
The chaotic test framework perform large amounts of random operations to test on the correctness and performance.

After setting up lock servers, in Tribbler repo run the following commands for purpose of testing.

This is the command to run performance stress test on read write throughput with random backend node join and leave.
> cargo test --package lab --test final_project_test_and_evaluation -- test_bin_storage_failure --exact --nocapture

This is the command to run performance stress test on read write throughput with no backend node failure
> cargo test --package lab --test final_project_test_and_evaluation -- test_bin_storage_no_failure --exact --nocapture

This is the command to run chaotic correctness test with random backend node join and leave.
> cargo test --package lab --test final_project_chaotic_correctness_test -- test_random_operations --exact --nocapture