#!/bin/bash

function kubeutl () (
  # Collection of a few short-handed commands for kubernetes stuff
  # works with oc, if you're into that
  # not sure where I'm going with this, tbh
  # TODO - Change quit behaviour to not change anything

  FUNC="$1"
  FUNC_ARG="$2"

  function chns () {
    ### Changes namespace in the current context ###
    # See if there is a $1
    if [[ ! -z "$1" ]]; then
      local NAMESPACE_INPUT="$1"
    else 
      #https://askubuntu.com/questions/1705/how-can-i-create-a-select-menu-in-a-shell-script
      echo "Here are a list of namespaces in the current context. Choose one."
      # Aggregate a list of namespaces in the current context
      #https://kubernetes.io/docs/reference/kubectl/overview/#custom-columns
      select opt in \
        $(kubectl get namespaces --no-headers=true -o custom-columns=NAME:.metadata.name) \
        "Quit"; 
      do
        # Assign set the desired namespace if the option isn't 'Quit'
        case "$opt" in
          "Quit")
            echo "You dediced to quit for some reason. Exiting..."
            user_quit=true
            break;;
          *)
            local NAMESPACE_INPUT="$opt"
            break;;
        esac
      done
    fi

    if [[ $user_quit ]]; then
      exit
    else 
      local CURRENT_CONTEXT="$(kubectl config current-context)"
      echo "Changing your current context to '$NAMESPACE_INPUT'."
      kubectl config set-context $CURRENT_CONTEXT --namespace=$NAMESPACE_INPUT
    fi
  }
  function chctx () {
    ### Changes context based on what's in $KUBECONFIG ###
    if [[ ! -z "$1" ]]; then
      local CONTEXT_INPUT="$1"
    else 
      echo "Here are a list of contexts Choose one."
      # Aggregate a list of contexts
      select opt in \
        $(kubectl config get-contexts --no-headers=true --output=name) \
        "Quit"; 
      do
        case "$opt" in
          "Quit")
            echo "You dediced to quit for some reason. Exiting..."
            user_quit=true
            break;;
          *)
            local CONTEXT_INPUT="$opt"
            break;;
        esac
      done
    fi

    if [[ $user_quit ]]; then
      exit
    else 
      echo "Changing your current context to '$CONTEXT_INPUT'."
      kubectl config use-context "$CONTEXT_INPUT"
    fi
  }

  function lspods () {
    ### Lists pods within all namespaces ###
    # '--all' can be passed to list in all namespaces
    kubectl get pods -o=jsonpath='{range .items[*]}{.metadata.namespace}:{.metadata.name}{"\n"}{end}' --all-namespaces 2> /dev/null
  }

  function nodetop () {
    ### Resource useage at the pod level per named node ###
    # TODO - Use jsonpath if possible
    echo "Not yet working"
    exit
    #regex="(Non-terminated\sPods\:\s)(.*?)(Allocated\sresources\:\s)" Gen 1
    regex="(Non-terminated\sPods\:\s)(.*?)(Allocated\sresources\:\s)" # Gen 2

    #sed_regex="([a-z]\{1,\})"
    # Namespace and pod
    # Cpu request and limit
    #sed_regex+="\s"
    # Memory request and limit
    #sed_regex+="\s"



    kube_output="$(kubectl describe nodes 2> /dev/null)"
    #echo 'not finished'
    if [[ $? -eq 0 && $kube_output =~ $regex ]]; then
      raw_text="${BASH_REMATCH[2]}" 
      echo $raw_text | sed -r "s/$sed_regex/\1\n/g"
    else
      echo "Something went wrong"
    fi

    #kubectl describe nodes | grep -Ezo "(Non-terminated\sPods\:\s)(.*?)(Allocated\sresources\:)"
    #kubectl describe nodes | grep "Allocated resources:" -A 4 --color=auto
  }

  select action in \
    "chns" \
    "chctx" \
    "lspods" \
    "nodetop"; 
    do
      case $action in 
        "chns")
          chns
          break;;
        "chctx")
          chctx
          break;;
        "lspods")
          lspods
          break;;
        "nodetop")
          nodetop
          break;;
        *)
      esac
    done
)
