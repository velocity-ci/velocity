module Page.Project.Commits exposing (..)

import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick, on, targetValue)
import Data.Commit as Commit exposing (Commit)
import Data.CommitResults as CommitResults exposing (Results)
import Data.Session as Session exposing (Session)
import Data.Project as Project exposing (Project)
import Data.Branch as Branch exposing (Branch)
import Page.Errored as Errored exposing (PageLoadError, pageLoadError)
import Page.Helpers exposing (formatDate, formatTime, sortByDatetime)
import Request.Project
import Util exposing ((=>))
import Task exposing (Task)
import Views.Page as Page
import Http
import Dict exposing (Dict)
import Time.DateTime as DateTime exposing (DateTime)
import Time.Date as Date exposing (Date)
import Page.Helpers exposing (formatDate)
import Route exposing (Route)
import Page.Project.Route as ProjectRoute
import Json.Decode as Decode


-- MODEL --


type alias Model =
    { commits : List Commit
    , total : Int
    , submitting : Bool
    , branch : Maybe Branch
    }


init : Session -> Project.Id -> Maybe Branch -> Task PageLoadError Model
init session id maybeBranch =
    let
        maybeAuthToken =
            Maybe.map .token session.user

        loadCommits =
            maybeAuthToken
                |> Request.Project.commits id maybeBranch
                |> Http.toTask

        initialModel { results, total } =
            { commits = results
            , total = total
            , submitting = False
            , branch = maybeBranch
            }

        handleLoadError _ =
            pageLoadError Page.Project "Project unavailable."
    in
        Task.map initialModel loadCommits
            |> Task.mapError handleLoadError



-- VIEW --


commitListToDict : List Commit -> Dict ( Int, Int, Int ) (List Commit)
commitListToDict commits =
    let
        reducer commit dict =
            let
                date =
                    commit.date
                        |> DateTime.date
                        |> Date.toTuple

                insert =
                    case Dict.get date dict of
                        Just exists ->
                            commit :: exists

                        Nothing ->
                            [ commit ]
            in
                Dict.insert date insert dict
    in
        List.foldl reducer Dict.empty commits


view : Project -> List Branch -> Model -> Html Msg
view project branches model =
    let
        commits =
            commitListToDict model.commits
                |> viewCommitListContainer project
    in
        div []
            [ viewCommitToolbar model.branch branches
            , commits
            ]


viewCommitToolbar : Maybe Branch -> List Branch -> Html Msg
viewCommitToolbar selectedBranch branches =
    let
        o b =
            option
                [ selected (b == selectedBranch) ]
                [ text (Branch.nameToString b) ]

        branchesSelect =
            List.map Just branches
                |> List.append [ Nothing ]
                |> List.map o
                |> select [ class "form-control", onChange ]

        onChange =
            on "change" (Decode.map FilterBranch Branch.selectDecoder)
    in
        nav [ class "navbar bg-light" ]
            [ branchesSelect ]


viewCommitListContainer : Project -> Dict ( Int, Int, Int ) (List Commit) -> Html Msg
viewCommitListContainer project dict =
    let
        listItemToDate dateListItem =
            dateListItem
                |> Tuple.first
                |> Date.fromTuple

        sortDateList a b =
            listItemToDate a
                |> Date.compare (listItemToDate b)
    in
        dict
            |> Dict.toList
            |> List.sortWith sortDateList
            |> List.map (viewCommitList project)
            |> div [ class "mt-3" ]


viewCommitList : Project -> ( ( Int, Int, Int ), List Commit ) -> Html Msg
viewCommitList project ( dateTuple, commits ) =
    let
        commitListItems =
            sortByDatetime .date commits
                |> List.map (viewCommitListItem project.id)

        formattedDate =
            Date.fromTuple dateTuple
                |> formatDate
    in
        div [ class "default-margin-bottom" ]
            [ h6 [ class "mb-2 text-muted" ] [ text formattedDate ]
            , div [ class "card" ]
                [ div [ class "list-group list-group-flush" ] commitListItems
                ]
            ]


viewCommitListItem : Project.Id -> Commit -> Html Msg
viewCommitListItem id commit =
    let
        truncatedHash =
            Commit.truncateHash commit.hash

        route =
            Route.Project id <| ProjectRoute.Commit commit.hash
    in
        a [ class "list-group-item list-group-item-action flex-column align-items-start", Route.href route ]
            [ div [ class "d-flex w-100 justify-content-between" ]
                [ h5 [ class "mb-1 text-overflow" ] [ text commit.message ]
                , small [] [ text truncatedHash ]
                ]
            , small [] [ strong [] [ text commit.author ], text (" commited at " ++ formatTime commit.date) ]
            ]


breadcrumb : Project -> List ( Route, String )
breadcrumb project =
    [ ( Route.Project project.id (ProjectRoute.Commits Nothing), "Commits" ) ]


viewBreadcrumbExtraItems : Model -> Html Msg
viewBreadcrumbExtraItems model =
    div [ class "ml-auto p-2" ]
        [ button
            [ class "ml-auto btn btn-dark", type_ "button", onClick SubmitSync, disabled model.submitting ]
            [ i [ class "fa fa-refresh" ] [], text " Refresh " ]
        ]



--pagination : Int -> Feed -> FeedSource -> Html Msg
--pagination activePage feed feedSource =
--    let
--        articlesPerPage =
--            limit feedSource
--
--        totalPages =
--            ceiling (toFloat feed.articlesCount / toFloat articlesPerPage)
--    in
--        if totalPages > 1 then
--            List.range 1 totalPages
--                |> List.map (\page -> pageLink page (page == activePage))
--                |> ul [ class "pagination" ]
--        else
--            Html.text ""
--
--
--pageLink : Int -> Bool -> Html Msg
--pageLink page isActive =
--    li [ classList [ "page-item" => True, "active" => isActive ] ]
--        [ a
--            [ class "page-link"
--            , href "javascript:void(0);"
--            , onClick (SelectPage page)
--            ]
--            [ text (toString page) ]
--        ]
-- UPDATE --


type Msg
    = SubmitSync
    | SyncCompleted (Result Http.Error Results)
    | FilterBranch (Maybe Branch.Name)


update : Project -> Session -> Msg -> Model -> ( Model, Cmd Msg )
update project session msg model =
    case msg of
        SubmitSync ->
            let
                getCommits authToken =
                    Request.Project.commits project.id model.branch (Just authToken)
                        |> Http.toTask

                cmdFromAuth authToken =
                    authToken
                        |> Request.Project.sync project.id
                        |> Http.toTask
                        |> Task.andThen (getCommits authToken |> always)
                        |> Task.attempt SyncCompleted

                cmd =
                    session
                        |> Session.attempt "sync project" cmdFromAuth
                        |> Tuple.second
            in
                { model | submitting = True } => cmd

        SyncCompleted (Ok { results, total }) ->
            { model
                | submitting = False
                , commits = results
                , total = total
            }
                => Cmd.none

        SyncCompleted (Err err) ->
            { model | submitting = False } => Cmd.none

        FilterBranch maybeBranch ->
            let
                uriEncoded =
                    maybeBranch
                        |> Maybe.andThen
                            (\(Branch.Name slug) ->
                                slug
                                    |> Http.encodeUri
                                    |> Branch.Name
                                    |> Just
                            )

                newRoute =
                    Route.Project project.id <| ProjectRoute.Commits uriEncoded
            in
                model => Route.modifyUrl newRoute
