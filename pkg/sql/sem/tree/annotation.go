package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

type AnnotationIdx int32

const NoAnnotation AnnotationIdx = 0

type AnnotatedNode struct {
	AnnIdx AnnotationIdx
}

func (n AnnotatedNode) GetAnnotation(ann *Annotations) interface{} {
	__antithesis_instrumentation__.Notify(603251)
	if n.AnnIdx == NoAnnotation {
		__antithesis_instrumentation__.Notify(603253)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(603254)
	}
	__antithesis_instrumentation__.Notify(603252)
	return ann.Get(n.AnnIdx)
}

func (n AnnotatedNode) SetAnnotation(ann *Annotations, annotation interface{}) {
	__antithesis_instrumentation__.Notify(603255)
	ann.Set(n.AnnIdx, annotation)
}

type Annotations []interface{}

func MakeAnnotations(numAnnotations AnnotationIdx) Annotations {
	__antithesis_instrumentation__.Notify(603256)
	return make(Annotations, numAnnotations)
}

func (a *Annotations) Set(idx AnnotationIdx, annotation interface{}) {
	__antithesis_instrumentation__.Notify(603257)
	(*a)[idx-1] = annotation
}

func (a *Annotations) Get(idx AnnotationIdx) interface{} {
	__antithesis_instrumentation__.Notify(603258)
	return (*a)[idx-1]
}
