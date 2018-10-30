module Page.Project.Commits exposing (Model, Msg(..), branchFilterConfig, branchFilterContext, breadcrumb, channelName, commitListToDict, events, init, loadCommits, pageLink, pagination, perPage, refreshButton, subscriptions, update, view, viewBreadcrumbExtraItems, viewCommitList, viewCommitListContainer, viewCommitListItem)

import Bootstrap.Button as Button
import Component.DropdownFilter as DropdownFilter
import Context exposing (Context)
import Css exposing (..)
import Data.AuthToken as AuthToken exposing (AuthToken)
import Data.Branch as Branch exposing (Branch)
import Data.Commit as Commit exposing (Commit)
import Data.Device as Device
import Data.PaginatedList as PaginatedList exposing (Paginated(..), PaginatedList)
import Data.Project as Project exposing (Project)
import Data.Session as Session exposing (Session)
import Dict exposing (Dict)
import Dom
import Html exposing (Html)
import Html.Attributes as UnstyledAttribute
import Html.Events as UnstyledEvents
import Html.Styled as Styled exposing (..)
import Html.Styled.Attributes as StyledAttribute exposing (class, classList, css)
import Html.Styled.Events exposing (on, onClick, targetValue)
import Http
import Json.Encode as Encode
import Navigation
import Page.Errored as Errored exposing (PageLoadError, pageLoadError)
import Page.Helpers exposing (formatDate, formatTime, sortByDatetime)
import Page.Project.Commit.Route as CommitRoute
import Page.Project.Route as ProjectRoute
import Request.Commit
import Request.Errors
import Request.Project
import Route exposing (Route)
import Task exposing (Task)
import Time.Date as Date exposing (Date)
import Time.DateTime as DateTime exposing (DateTime)
import Util exposing ((=>))
import Views.Commit exposing (branchList, commitTimeInformation)
import Views.Helpers exposing (styledOnClickPage)
import Views.Page as Page
import Views.Style as Style


-- MODEL --


type alias Model =
    { commits : List Commit
    , total : Int
    , page : Int
    , submitting : Bool
    , branch : Maybe Branch
    , dropdownState : DropdownFilter.DropdownState
    , branchFilterTerm : String
    }


loadCommits : Context -> Project.Slug -> Maybe AuthToken -> Maybe Branch -> Int -> Task Request.Errors.HttpError (PaginatedList Commit)
loadCommits context projectSlug maybeAuthToken maybeBranch page =
    maybeAuthToken
        |> Request.Commit.list context projectSlug (Maybe.map .name maybeBranch) perPage page


init : Context -> Session msg -> List Branch -> Project -> Maybe Branch.Name -> Maybe Int -> Task PageLoadError Model
init context session branches project maybeBranchName maybePage =
    let
        defaultPage =
            Maybe.withDefault 1 maybePage

        maybeAuthToken =
            Maybe.map .token session.user

        maybeBranch =
            branches
                |> List.filter (\b -> maybeBranchName == Just b.name)
                |> List.head

        initialModel commits =
            { commits = PaginatedList.results commits
            , total = PaginatedList.total commits
            , page = defaultPage
            , submitting = False
            , branch = maybeBranch
            , dropdownState = DropdownFilter.initialDropdownState
            , branchFilterTerm = ""
            }

        handleLoadError _ =
            pageLoadError Page.Project "Project unavailable."
    in
        Task.map initialModel (loadCommits context project.slug maybeAuthToken maybeBranch defaultPage)
            |> Task.andThen
                (\initialModel ->
                    case ( maybeAuthToken, project.synchronising ) of
                        ( Just authToken, False ) ->
                            Request.Project.sync context project.slug authToken
                                |> Task.andThen (\_ -> Task.succeed initialModel)

                        _ ->
                            Task.succeed initialModel
                )
            |> Task.mapError handleLoadError


perPage : Int
perPage =
    10


branchFilterConfig : DropdownFilter.Config Msg Branch
branchFilterConfig =
    { dropdownMsg = BranchFilterDropdownMsg
    , termMsg = BranchFilterTermMsg
    , noOpMsg = NoOp
    , selectItemMsg = FilterBranch
    , labelFn = .name >> Just >> Branch.nameToString
    , icon = toUnstyled <| i [ class "fa fa-code-fork" ] []
    , showFilter = True
    , showAllItemsItem = True
    }


branchFilterContext : Device.Size -> List Branch -> Model -> DropdownFilter.Context Branch
branchFilterContext deviceSize branches { dropdownState, branchFilterTerm, branch } =
    let
        items =
            branches
                |> List.filter .active
                |> List.sortBy branchFilterConfig.labelFn
    in
        { items = items
        , dropdownState = dropdownState
        , filterTerm = branchFilterTerm
        , selectedItem = branch
        , deviceSize = deviceSize
        }



-- SUBSCRIPTIONS --


subscriptions : Device.Size -> List Branch -> Model -> Sub Msg
subscriptions deviceSize branches model =
    branchFilterContext deviceSize branches model
        |> DropdownFilter.subscriptions branchFilterConfig



-- CHANNELS --


channelName : Project.Slug -> String
channelName projectSlug =
    "project:" ++ Project.slugToString projectSlug


events : Project.Slug -> Dict String (List ( String, Encode.Value -> Msg ))
events projectSlug =
    let
        pageEvents =
            [ ( "commit:new", always RefreshCommitList )
            , ( "commit:update", always RefreshCommitList )
            , ( "commit:deleted", always RefreshCommitList )
            ]
    in
        Dict.singleton (channelName projectSlug) pageEvents



-- VIEW --


view : Project -> List Branch -> Device.Size -> Model -> Html.Html Msg
view project branches deviceSize model =
    let
        commits =
            commitListToDict model.commits
                |> viewCommitListContainer project

        branchFilter =
            branchFilterContext deviceSize branches model
                |> DropdownFilter.view branchFilterConfig
                |> fromUnstyled

        refreshCommitsButton =
            refreshButton project model

        buttons =
            div [ class "btn-toolbar" ]
                [ branchFilter
                , refreshCommitsButton
                ]

        paginationToolbar =
            div [ class "btn-toolbar" ]
                [ pagination model.page model.total project model.branch ]
    in
        div []
            [ h4 [ class "mb-2" ] [ text "Commits" ]
            , buttons
            , commits
            , paginationToolbar
            ]
            |> toUnstyled


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


viewCommitListContainer : Project -> Dict ( Int, Int, Int ) (List Commit) -> Styled.Html Msg
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
            |> List.take perPage
            |> List.map (viewCommitList project)
            |> div [ class "mt-3" ]


viewCommitList : Project -> ( ( Int, Int, Int ), List Commit ) -> Styled.Html Msg
viewCommitList project ( dateTuple, commits ) =
    let
        commitListItems =
            sortByDatetime .date commits
                |> List.map (viewCommitListItem project.slug)

        formattedDate =
            Date.fromTuple dateTuple
                |> formatDate
    in
        div []
            [ h6 [ class "mb-2 mt-2 text-muted" ] [ text formattedDate ]
            , div [ class "card" ]
                [ div [ class "list-group list-group-flush" ] commitListItems
                ]
            ]


viewCommitListItem : Project.Slug -> Commit -> Styled.Html Msg
viewCommitListItem slug commit =
    let
        truncatedHash =
            Commit.truncateHash commit.hash

        route =
            Route.Project slug <| ProjectRoute.Commit commit.hash CommitRoute.Overview
    in
        a
            [ class "list-group-item list-group-item-action flex-column align-items-start"
            , Route.styledHref route
            , styledOnClickPage NewUrl route
            ]
            [ div [ class "d-flex w-100 justify-content-between" ]
                [ h6
                    [ class "mb-1"
                    , css [ Style.textOverflowMixin ]
                    ]
                    [ text commit.message ]
                , Styled.small [] [ text truncatedHash ]
                ]
            , div [ class "d-flex w-100 justify-content-between" ]
                [ fromUnstyled <| commitTimeInformation commit
                , fromUnstyled <| branchList commit
                ]
            ]


breadcrumb : Project -> List ( Route, String )
breadcrumb project =
    [ ( Route.Project project.slug (ProjectRoute.Commits Nothing Nothing), "Commits" ) ]


viewBreadcrumbExtraItems : Project -> Model -> Html.Html Msg
viewBreadcrumbExtraItems project model =
    toUnstyled <| text ""


refreshButton : Project -> Model -> Styled.Html Msg
refreshButton project model =
    let
        submitting =
            project.synchronising || model.submitting

        iconClassList =
            [ "fa fa-refresh" => True
            , "fa-spin fa-fw" => submitting
            ]
    in
        Button.button
            [ Button.outlineSecondary
            , Button.attrs
                [ UnstyledAttribute.class "ml-auto btn btn-dark"
                , UnstyledEvents.onClick SubmitSync
                , UnstyledAttribute.disabled submitting
                ]
            ]
            [ toUnstyled (i [ classList iconClassList ] []) ]
            |> fromUnstyled


pagination : Int -> Int -> Project -> Maybe Branch -> Styled.Html Msg
pagination activePage total project maybeBranch =
    let
        totalPages =
            ceiling (toFloat total / toFloat perPage)
    in
        if totalPages > 1 then
            List.range 1 totalPages
                |> List.map (\page -> pageLink page (page == activePage) project maybeBranch)
                |> ul [ class "pagination mt-3" ]
        else
            text ""


pageLink : Int -> Bool -> Project -> Maybe Branch -> Styled.Html Msg
pageLink page isActive project maybeBranch =
    let
        route =
            Route.Project project.slug <| ProjectRoute.Commits (Maybe.map .name maybeBranch) (Just page)
    in
        li [ classList [ "page-item" => True, "active" => isActive ] ]
            [ a
                [ class "page-link"
                , Route.styledHref route
                , styledOnClickPage NewUrl route
                ]
                [ text (toString page) ]
            ]



-- UPDATE --


type Msg
    = SubmitSync
    | SyncCompleted (Result Request.Errors.HttpError Project)
    | FilterBranch (Maybe Branch)
    | SelectPage Int
    | NewUrl String
    | RefreshCommitList
    | RefreshCompleted (Result Request.Errors.HttpError (PaginatedList Commit))
    | BranchFilterDropdownMsg DropdownFilter.DropdownState
    | BranchFilterTermMsg String
    | NoOp


update : Context -> Project -> Session msg -> Msg -> Model -> ( Model, Cmd Msg )
update context project session msg model =
    case msg of
        NewUrl newUrl ->
            model => Navigation.newUrl newUrl

        SubmitSync ->
            let
                cmdFromAuth authToken =
                    authToken
                        |> Request.Project.sync context project.slug
                        |> Task.attempt SyncCompleted

                cmd =
                    session
                        |> Session.attempt "sync project" cmdFromAuth
                        |> Tuple.second
            in
                { model | submitting = True } => cmd

        SyncCompleted (Ok _) ->
            { model | submitting = False }
                => Cmd.none

        SyncCompleted (Err _) ->
            { model | submitting = False }
                => Cmd.none

        SelectPage page ->
            let
                uriEncoded =
                    model.branch
                        |> Maybe.map .name
                        |> Maybe.andThen
                            (\(Branch.Name slug) ->
                                slug
                                    |> Http.encodeUri
                                    |> Branch.Name
                                    |> Just
                            )

                newRoute =
                    Route.Project project.slug <| ProjectRoute.Commits uriEncoded (Just page)
            in
                model => Route.modifyUrl newRoute

        FilterBranch maybeBranch ->
            let
                uriEncoded =
                    maybeBranch
                        |> Maybe.map .name
                        |> Maybe.andThen
                            (\(Branch.Name slug) ->
                                slug
                                    |> Http.encodeUri
                                    |> Branch.Name
                                    |> Just
                            )

                newRoute =
                    Route.Project project.slug <| ProjectRoute.Commits uriEncoded (Just 1)
            in
                { model | commits = [] }
                    => Route.modifyUrl newRoute

        RefreshCommitList ->
            let
                refreshTask =
                    loadCommits context project.slug (Maybe.map .token session.user) model.branch model.page
            in
                model => Task.attempt RefreshCompleted refreshTask

        RefreshCompleted (Ok paginatedResults) ->
            { model
                | commits = PaginatedList.results paginatedResults
                , total = PaginatedList.total paginatedResults
            }
                => Cmd.none

        RefreshCompleted (Err _) ->
            model => Cmd.none

        BranchFilterDropdownMsg state ->
            { model | dropdownState = state }
                => Task.attempt (always NoOp) (Dom.focus "filter-item-input")

        BranchFilterTermMsg term ->
            { model | branchFilterTerm = term }
                => Cmd.none

        NoOp ->
            model => Cmd.none
