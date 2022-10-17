import { getReplicas } from "api/api"
import { Replica } from "api/typesGenerated"
import { createMachine, assign } from "xstate"

export const highAvailabilityMachine = createMachine(
  {
    id: "highAvailabilityMachine",
    initial: "loaded",
    schema: {
      context: {} as {
        replicas?: Replica[]
        replicasError?: unknown
      },
      events: {} as { type: "REFRESH_REPLICAS" },
      services: {} as {
        getReplicas: {
          data: Replica[]
        }
      },
    },
    tsTypes: {} as import("./highAvailabilityMachine.typegen").Typegen0,
    states: {
      refreshReplicas: {
        invoke: {
          src: "getReplicas",
          onDone: {
            target: "loaded",
            actions: ["assignReplicas"],
          },
          onError: {
            target: "loaded",
            actions: ["assignReplicasError"],
          },
        },
      },
      loaded: {
        initial: "refreshingReplicas",
        states: {
          refreshingReplicas: {
            invoke: {
              id: "refreshReplicas",
              src: "getReplicas",
              onDone: { target: "waiting", actions: "assignReplicas" },
            },
          },
          waiting: {
            after: {
              5000: "refreshingReplicas",
            },
          },
        },
      },
    },
  },
  {
    services: {
      getReplicas,
    },
    actions: {
      assignReplicas: assign({
        replicas: (_, { data }) => data,
      }),
      assignReplicasError: assign({
        replicasError: (_, { data }) => data,
      }),
    },
  },
)
