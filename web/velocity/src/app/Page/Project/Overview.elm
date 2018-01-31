module Page.Project.Overview exposing (..)

import Html exposing (..)
import Html.Attributes exposing (..)
import Data.Project as Project exposing (Project)
import Page.Helpers exposing (formatDateTime)
import Data.Build as Build exposing (Build)
import Views.Build exposing (viewBuildStatusIcon, viewBuildTextClass)
import Util exposing ((=>))
import Data.Session as Session exposing (Session)
import Views.Helpers exposing (onClickPage)
import Navigation
import Route


-- MODEL --


type alias Model =
    {}


initialModel : Model
initialModel =
    {}



-- VIEW --


view : Project -> List Build -> Html msg
view project builds =
    div []
        [ viewOverviewCard project
        , viewBuildHistoryTable builds
        ]


viewOverviewCard : Project -> Html msg
viewOverviewCard project =
    div [ class "card mb-3" ]
        [ div [ class "card-body" ]
            [ dl [ class "mb-0" ]
                [ dt [] [ text "Repository" ]
                , dd [] [ text project.repository ]
                , dt [] [ text "Last update" ]
                , dd [] [ text (formatDateTime project.updatedAt) ]
                ]
            ]
        ]


viewBuildHistoryTable : List Build -> Html msg
viewBuildHistoryTable builds =
    div [ class "card" ]
        [ h5 [ class "card-header border-bottom-0" ] [ text "Build history" ]
        , table [ class "table mb-0" ] (List.map viewBuildHistoryTableRow (List.take 10 builds))
        ]


viewBuildHistoryTableRow : Build -> Html msg
viewBuildHistoryTableRow build =
    let
        rowClasses =
            [ (viewBuildTextClass build) => True ]
    in
        tr [ classList rowClasses ]
            [ td [] [ text (formatDateTime build.createdAt) ]
            , td [] [ viewBuildStatusIcon build ]
            ]



-- UPDATE --


type Msg
    = NewUrl String


update : Project -> Session msg -> Msg -> Model -> ( Model, Cmd Msg )
update project session msg model =
    case msg of
        NewUrl url ->
            model => Navigation.newUrl url
