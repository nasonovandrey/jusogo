

package source
import (
    "time"
)

const baseTimeInSeconds = 5
const HEARTBEAT = baseTimeInSeconds * time.Second
const TIMEOUT = (baseTimeInSeconds * 11 / 10) * time.Second
