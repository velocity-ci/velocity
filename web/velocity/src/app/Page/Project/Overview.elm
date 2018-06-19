module Page.Project.Overview exposing (..)

import Html exposing (..)
import Html.Attributes exposing (..)
import Data.Project as Project exposing (Project)
import Data.Build as Build exposing (Build)
import Util exposing ((=>))
import Data.Session as Session exposing (Session)
import Navigation
import Route
import Page.Project.Route as ProjectRoute
import Page.Project.Commit.Route as CommitRoute
import Views.Helpers exposing (onClickPage)
import Data.Task as Task
import Data.Commit as Commit
import Views.Build exposing (viewBuildHistoryTable)
import Page.Helpers exposing (formatDateTime)


-- MODEL --


type alias Model =
    {}


initialModel : Model
initialModel =
    {}



-- VIEW --


view : Project -> List Build -> Html Msg
view project builds =
    div []
        [ viewOverviewCard project
        , viewBuildHistoryTable project builds NewUrl
        ]


viewOverviewCard : Project -> Html Msg
viewOverviewCard project =
    div [ class "col-md-12 px-0 mx-0" ]
        [ dl [ class "mt-4 mb-5" ]
            [ dt [] [ text "Repository" ]
            , dd [] [ text project.repository ]
            , dt [] [ text "Last update" ]
            , dd [] [ text (formatDateTime project.updatedAt) ]
            ]
        ]



-- UPDATE --


type Msg
    = NewUrl String


update : Project -> Session msg -> Msg -> Model -> ( Model, Cmd Msg )
update project session msg model =
    case msg of
        NewUrl url ->
            model => Navigation.newUrl url
