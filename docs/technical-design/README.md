---
sidebar: auto
---

# Technical Design

This technical design document is currently in rapid development and unstable. It will be updated as frequently as possible but may often fall behind progress.

## Architect

"supervises construction" of builds

## Plugin

Event-based steps (these won't fail the build if they error): e.g.:

Build status notifications (e.g. Slack)
Build status hooks (e.g. commit status on GitHub/Gitlab)
These are only used by the architect and will be configured through global architect configuration. For our SaaS, we can provide a way for users to set global configuration for their Team/Company/Organisation.

We need to provide a way of knowing when these plugins error (maybe just add to our event log)

### Events

- `BUILD_START`
- `TASK_START`
- `BUILD_COMPLETE`
- `TASK_COMPLETE`
- `BUILD_SUCCESS`
- `TASK_SUCCESS`
- `BUILD_FAIL`
- `TASK_FAIL`

## Builder

Builds _Tasks_ that an architect gives it

## Blueprint

Task configuration, stored in yaml format e.g. https://github.com/velocity-ci/velocity/blob/master/tasks/backend/cli/publish.yml
A _Task_ is created from this

### Step

## Task

A build-time _Blueprint_ w/ any further configuration/plugins from the root `.velocity.yaml`. The _Architect_ schedules tasks on _Builders_

### Lifecycle

1. `waiting`: not ready for building as it is waiting for previous tasks (in stages) to finish.
2. `ready`: ready for building, thus can be scheduled to builders.
3. `scheduled`: sent to a builder.
4. `building`: building in progress on a builder. Returns to `ready` state if interrupted.
5. `succeeded`/`failed` final completed state.

### Step

#### Lifecycle

## Stage

A set of _Tasks_ that can run in parallel

### Lifecycle

1. `waiting`: not ready for building as it is waiting for previous stages to finish.
2. `building`: tasks from this stage are being scheduled to builders.
3. `succeeded`/`failed`: final completed state.

## Construction Plan

A collection of _Stages_ to be built in order. This allows for parallelization and ordering of tasks for a build, kind of like a Gantt chart. It will let us implement future things like _Blueprint_ dependencies and _Blueprint_ composition.

## Build

A runtime/built _ConstructionPlan_

### Lifecycle

1. `building`: A build is in progress.
2. `succeeded`/`failed`: final completed state.

When a user submits a build:

- A construction plan is generated, and the build, stages and tasks from that plan are persisted.
- Tasks from the first stage of the build are scheduled.
- When a task finishes:
  - if other tasks in the stage are successful and there is a next stage:
    - schedule tasks from the next stage.
  - else:
    - promote least successful task status to build status.

If a task is interrupted, it needs to be rescheduled.
