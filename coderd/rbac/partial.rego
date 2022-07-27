package authz
# opa eval --partial --format=pretty 'data.authz.allow = true' -d partial.rego --unknowns input.object.owner --unknowns input.object.org_owner -i input.json

import future.keywords

# Admins can read everything
# Org members can read everything in their org
# Members can read everything that is theirs

# SQL has
# - owner
# - org
#
#
# Site:
# WHERE true AND false
#
# Org:
# WHERE org = ? AND org != ?
#
# User
# WHERE owner = ? AND owner != ?

# Unknowns:
# object.owner
# object.org_owner


# bool_flip lets you assign a value to an inverted bool.
# You cannot do 'x := !false', but you can do 'x := bool_flip(false)'
bool_flip(b) = flipped {
    b
    flipped = false
}

bool_flip(b) = flipped {
    not b
    flipped = true
}

number(set) = c {
	count(set) == 0
    c := 0
}

number(set) = c {
	false in set
    c := -1
}

number(set) = c {
	not false in set
	set[_]
    c := 1
}


default site = 0
site := num {
	# relevent are all the permissions that affect the given unknown object
	allow := { x |
    	perm := input.subject.roles[_].site[_]
        perm.action in [input.action, "*"]
		perm.resource_type in [input.object.type, "*"]
        x := bool_flip(perm.negate)
    }
    num := number(allow)
}

org_members := { orgID |
	input.subject.roles[_].org[orgID]
}

default org = 0
org := num {
	orgPerms := { id: num |
		id := org_members[_]
		set := { x |
			perm := input.subject.roles[_].org[id][_]
			perm.action in [input.action, "*"]
			perm.resource_type in [input.object.type, "*"]
			x := bool_flip(perm.negate)
		}
		num := number(set)
	}

	num := orgPerms[input.object.org_owner]
}

org_mem := 1 {
	input.object.org_owner in org_members
}

#org := num {
#	not input.object.org_owner in org_members
#	num := -1
#}

default user = 0
user := num {
    input.subject.id = input.object.owner
	# relevent are all the permissions that affect the given unknown object
	allow := { x |
    	perm := input.subject.roles[_].user[_]
        perm.action in [input.action, "*"]
		perm.resource_type in [input.object.type, "*"]
        x := bool_flip(perm.negate)
    }
    num := number(allow)
}

default allow = false
# Site
allow {
	site = 1
}

# Org
allow {
	not site = -1
	org = 1
}

# User
allow {
	not site = -1
	not org = -1
	org_mem = 1
	user = 1
}

