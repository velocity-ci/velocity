module Page.Project exposing (Model, Msg, ViewConfiguration, id, init, update, view)

import Element exposing (..)
import Project exposing (Project)
import Project.Id exposing (Id)



-- Model


type alias Model =
    { id : Id }


init : Id -> ( Model, Cmd Msg )
init projectId =
    ( { id = projectId }, Cmd.none )



-- Info


id : Model -> Id
id model =
    model.id



-- Update


type Msg
    = NoOp


update : Msg -> Model -> ( Model, Cmd Msg )
update _ model =
    ( model, Cmd.none )



-- View


type alias ViewConfiguration =
    { model : Model
    , project : Project
    }


view : ViewConfiguration -> { title : String, content : Element msg }
view config =
    { title = "Project page"
    , content = text "Project page"
    }
