module Page.Build exposing (Model, Msg, init, subscriptions, toContext, toSession, update, view)

import Context exposing (Context)
import Element exposing (..)
import Project.Build as Build
import Project.Build.Id exposing (Id)
import Session exposing (Session)


-- Model


type alias Model msg =
    { session : Session msg
    , context : Context msg
    , id : Id
    }


init : Session msg -> Context msg -> Id -> ( Model msg, Cmd Msg )
init session context buildId =
    ( { session = session
      , context = context
      , id = buildId
      }
    , Cmd.none
    )



-- Subscriptions


subscriptions : Model msg -> Sub Msg
subscriptions model =
    Sub.none



-- Update


type Msg
    = NoOp


update : Msg -> Model msg -> ( Model msg, Cmd Msg )
update msg model =
    ( model
    , Cmd.none
    )



-- Export


toSession : Model msg -> Session msg
toSession model =
    model.session


toContext : Model msg -> Context msg
toContext model =
    model.context



-- View


view : Model msg -> { title : String, content : Element Msg }
view model =
    { title = "Project build page"
    , content = text "Project build page"
    }
