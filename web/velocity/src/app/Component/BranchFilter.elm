module Component.BranchFilter exposing (Context, Config, DropdownState, initialDropdownState, view, subscriptions)

{- A stateless BranchFilter component. -}
-- EXTERNAL --

import Html exposing (..)
import Html.Events exposing (onWithOptions, onInput)
import Html.Attributes exposing (..)
import Bootstrap.Dropdown as Dropdown
import Bootstrap.Button as Button
import Bootstrap.Form as Form
import Bootstrap.Form.Input as Input
import Json.Decode as Decode


-- INTERNAL --

import Util exposing ((=>))
import Data.Branch as Branch exposing (Branch)


-- MODEL --


type alias Context =
    { branches : List Branch
    , dropdownState : Dropdown.State
    , filterTerm : String
    , selectedBranch : Maybe Branch
    }


type alias Config msg =
    { dropdownMsg : Dropdown.State -> msg
    , termMsg : String -> msg
    , noOpMsg : msg
    , selectBranchMsg : Maybe Branch -> msg
    }


type alias DropdownState =
    Dropdown.State


initialDropdownState : DropdownState
initialDropdownState =
    Dropdown.initialState



-- SUBSCRIPTIONS --


subscriptions : Config msg -> Context -> Sub msg
subscriptions { dropdownMsg } { dropdownState } =
    Dropdown.subscriptions dropdownState dropdownMsg



-- VIEW --


view : Config msg -> Context -> Html msg
view config context =
    div []
        [ Dropdown.dropdown
            context.dropdownState
            { options = [ Dropdown.menuAttrs [ onClick (config.noOpMsg), class "branch-filter-dropdown" ] ]
            , toggleMsg = config.dropdownMsg
            , toggleButton = toggleButton context
            , items = viewDropdownItems config context
            }
        ]


toggleButton : Context -> Dropdown.DropdownToggle msg
toggleButton { selectedBranch } =
    let
        toggleText =
            selectedBranch
                |> Maybe.map .name
                |> Branch.nameToString
    in
        Dropdown.toggle
            [ Button.outlineSecondary ]
            [ i [ class "fa fa-code-fork" ] []
            , text (" " ++ toggleText)
            ]


viewDropdownItems : Config msg -> Context -> List (Dropdown.DropdownItem msg)
viewDropdownItems config context =
    let
        filterForm =
            Dropdown.customItem (viewForm config.termMsg context)

        branchItems =
            viewBranchItems config context

        allBranchItemButton =
            allBranchItem config context
    in
        filterForm :: Dropdown.divider :: allBranchItemButton :: Dropdown.divider :: branchItems


allBranchItem : Config msg -> Context -> Dropdown.DropdownItem msg
allBranchItem { selectBranchMsg } { selectedBranch } =
    branchItem selectBranchMsg selectedBranch Nothing


viewBranchItems : Config msg -> Context -> List (Dropdown.DropdownItem msg)
viewBranchItems config { branches, filterTerm, selectedBranch } =
    branches
        |> List.filter (branchFilter filterTerm)
        |> List.sortBy (.name >> Just >> Branch.nameToString)
        |> List.map (Just >> branchItem config.selectBranchMsg selectedBranch)


branchFilter : String -> Branch -> Bool
branchFilter filterTerm { name, active } =
    active && String.contains filterTerm (Branch.nameToString (Just name))


branchItem : (Maybe Branch -> msg) -> Maybe Branch -> Maybe Branch -> Dropdown.DropdownItem msg
branchItem selectMsg selectedBranch maybeBranch =
    let
        itemText =
            maybeBranch
                |> Maybe.map .name
                |> Branch.nameToString
                |> text

        itemIcon =
            if selectedBranch == maybeBranch then
                i [ class "fa fa-check" ] []
            else
                text ""
    in
        Dropdown.buttonItem
            [ onClick (selectMsg <| maybeBranch) ]
            [ itemIcon
            , itemText
            ]


viewForm : (String -> msg) -> Context -> Html msg
viewForm msg { filterTerm } =
    Form.form [ class "px-2 py-0", style [ "width" => "400px" ] ]
        [ Form.group []
            [ Input.email
                [ Input.id "filter-branch-input"
                , Input.placeholder "Filter branches"
                , Input.attrs [ onInput msg ]
                , Input.value filterTerm
                ]
            ]
        ]



-- helper to cancel click anywhere


onClick : msg -> Attribute msg
onClick message =
    onWithOptions
        "click"
        { stopPropagation = True
        , preventDefault = False
        }
        (Decode.succeed message)
