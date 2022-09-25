import NoVncClient from "@novnc/novnc/core/rfb"
import { FC, useLayoutEffect, useRef } from "react"

export const VNCPage: FC = () => {
  const rootRef = useRef<HTMLDivElement>(null)
  useLayoutEffect(() => {
    console.log("Root ref!", typeof rootRef.current, rootRef.current)
    if (!rootRef.current) {
      return
    }
    console.log("Creating client!")
    const client = new NoVncClient(rootRef.current, "ws://localhost:5901", {
        credentials: {
            username: "user",
            password: "alpine",
        } as any,
    })
    client.scaleViewport = true
    client.resizeSession = true
    client.clipViewport = true
    client.compressionLevel = 9
    client.qualityLevel = 9
    return () => {
        client.disconnect()
    }
  }, [rootRef])
  return (
    <div
      ref={rootRef}
      style={{
        width: "100vw",
        height: "100vh",
      }}
    />
  )
}
