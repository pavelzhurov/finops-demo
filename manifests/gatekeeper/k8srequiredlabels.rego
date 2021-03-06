package k8srequiredlabels

get_labels[labels] {
    input.review.object.kind == ["Pod"][_]
    labels := {key: value | value := input.review.object.metadata.labels[key]}
}

get_labels[labels] {
    input.review.object.kind == ["ReplicaSet", "Rollout"][_]
    labels := {key: value | value := input.review.object.spec.template.metadata.labels[key]}
}

what_is_wrong_with_labels[msg] {
    get_labels[labels]
    missing := {key | input.parameters.labels[key]} - {key | labels[key]}
    count(missing) > 0
    msg := sprintf("You didn't provide labels: %v", [missing])
}

what_is_wrong_with_labels[msg] {
    get_labels[labels]
    labels[key] != input.parameters.labels[key]
    value := labels[key]
    msg := sprintf("You have wrong values for key \"%v\": %v", [key, value])
}

violation[{"msg": msg}] {
    errors := [error | what_is_wrong_with_labels[error]]
    count(errors) > 0
    msg := concat("\n", errors)
}