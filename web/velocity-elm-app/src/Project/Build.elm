module Project.Build exposing (Build, createdAt)

import Project.Build.Id as Id exposing (Id)
import Project.Build.Status exposing (Status)
import Project.Build.Step exposing (Step)
import Project.Task as Task exposing (Task)
import Time


type Build
    = Build Internals


type alias Internals =
    { id : Id
    , status : Status
    , task : Task
    , steps : List Step
    , createdAt : Time.Posix
    , completedAt : Maybe Time.Posix
    , updatedAt : Maybe Time.Posix
    , startedAt : Maybe Time.Posix
    }



-- INFO


createdAt : Build -> Time.Posix
createdAt (Build rec) =
    rec.createdAt
