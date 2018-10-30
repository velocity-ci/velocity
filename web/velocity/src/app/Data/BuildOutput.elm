module Data.BuildOutput exposing (Step, TaskStep, joinSteps)

-- EXTERNAL --
-- INTERNAL --

import Array
import Data.BuildStep as BuildStep exposing (BuildStep)
import Data.Task as ProjectTask exposing (Parameter(..), Step(..))


type alias TaskStep =
    ProjectTask.Step


type alias Step =
    ( TaskStep, BuildStep )


joinSteps : ProjectTask.Task -> BuildStep -> Maybe Step
joinSteps task buildStep =
    task
        |> .steps
        |> Array.fromList
        |> Array.get buildStep.number
        |> Maybe.map (\taskStep -> ( taskStep, buildStep ))
