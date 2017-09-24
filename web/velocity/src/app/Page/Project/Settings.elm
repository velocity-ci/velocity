module Page.Project.Settings exposing (..)

import Html exposing (..)
import Data.Session as Session exposing (Session)
import Data.Project as Project exposing (Project)
import Util exposing ((=>))


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



-- UPDATE --


type Msg
    = NoOp


update : Session -> Msg -> Model -> ( Model, Cmd Msg )
update session msg model =
    model => Cmd.none
