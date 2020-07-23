# Motivation
Minio is a lightweight binary. It has been optimized to restart quickly and make the impact of binary restart be
non impactful to applications. 

For baremetal the recommended upgrade scheme has been to issue an update command in parallel to all instances of minio.

The benefits of such an approach are significant enough to try to replicate the same when running in kubernetes.
Unlike bare metal, restarting pods is a heavier operation. The following state machine attempts to achieve a non disruptive 
upgrade while minimizing the time window in which minio quorum needs to run in a mixed version mode.
 
# Basic process

1. Validate the container image
    1. Goals:
        1. Make sure the container image, and the binary are official
        2. Make sure this is an upgrade, downgrades could be achieved with an override flag in the CRD update.   
    2. Ensure the container image and tag as signed and newer than the existing image and above the minimum version for minio that supports this upgrade.
    3. Extract the binary locally within the operator pod and validate the binary is consistent with the container image.
    4. Failure:
        1. OperatorPodRestart: Pod checks if the container image is downloaded locally and resumes with verification
2. Ensure that pods do not restart due to changing of the container image in yaml.
    1. Goals:
        1. Make sure any pod restarts during upgrade start up with the newer version.
        2. Make sure rolling updates are not triggered during upgrade4
    3. Update the stateful set (atomic)
        1. Change `UpdateStrategy` to `OnDelete`
        2. Operator applies the label for new and old version of container image
        3. Operator applies the new container image to stateful set.
    4. Failures:
        1. OperatorPodRestart: Steps above are idempotent, operator can resume 2.
3. Operator detects the ongoing upgrade and triggers the API to have minio fetch the binary from the operator.
    1. Goals:
        1. Establish the step the operator resumes from on accidental operator pod restart
        2. Operator distributes the binary to all the minio pods.
    2. For each minio pod, operator checks the version of minio and if it is the older version continue else head to step 4.
        1. This is in case the operator reached here due to a restart
        2. If no pod is running the older version operator jumps to step 5
    3. Minio uses the Operator APIs to fetch the binary image
    4. Failures:
        1. OperatorPodRestart: Goes back to step 3
        2. Minio Pod Restart:
            1. New pod runs new version
            2. Either quorum will run the new version or old.
4. Operator issues the command to switch binary once all the minio pods have the new binary.
    1. Goals:
        1. Establish a step where all minio pods have the new binary and are ready to restart
        2. Issue the restart in parallel
    2. Minio validates the binary signature.
    3. Minio restarts with the new binary
    4. Failures:
        1. Minio Pod Restart:
            1. The new pod runs the new version.
            2. Either quorum will run the new version or old.
        2. OperatorPodRestart: Goes back to step 3
5. Operator changes back to `RollingUpdate`
    1. Goals:
        1. Check if the pod instances are all ready to be updated to the latest version of the stateful set.
        2. Issue the rolling update to upgrade the stateful set yaml. This is needed to update each pod instance to pick up the latest yaml. So far only the stateful set has the latest yaml not the pod instance of the stateful set.
    2. Operator checks the version of minio in all the pods.
    3. Operator updates stateful set (Final step)
        1. Remove upgrade labels
        d. changes `UpdateStrategy` to `RollingUpdate`. This triggers a rolling restart
    4. Failures:
        1. Minio Pod Restart:
            1. The new pod runs the new version.
            2. Either quorum will run the new version or old.
        2. OperatorPodRestart: Goes back to step 3, if stateful set has not been updated.
