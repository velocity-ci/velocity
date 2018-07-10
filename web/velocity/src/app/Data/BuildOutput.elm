module Data.BuildOutput exposing (..)

-- EXTERNAL --

import Array


-- INTERNAL --

import Data.Task as ProjectTask exposing (Step(..), Parameter(..))
import Data.BuildStep as BuildStep exposing (BuildStep)


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
