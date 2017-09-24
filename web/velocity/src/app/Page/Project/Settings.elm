module Page.Project.Settings exposing (..)

import Html exposing (..)
import Data.Session as Session exposing (Session)
import Data.Project as Project exposing (Project)
import Util exposing ((=>))
import Route exposing (Route)
import Page.Project.Route as ProjectRoute


-- MODEL --


type alias Model =
    {}


init : Project -> Model
init project =
    {}



-- VIEW --


view : Model -> Html Msg
view model =
    div [] [ text "Setting page" ]


breadcrumb : Project -> List ( Route, String )
breadcrumb project =
    [ ( Route.Project ProjectRoute.Settings project.id, "Settings" ) ]



-- UPDATE --


type Msg
    = NoOp


update : Session -> Msg -> Model -> ( Model, Cmd Msg )
update session msg model =
    model => Cmd.none
