package org.wso2.apim.grpc.telemetry.receiver.generated;

import io.grpc.stub.ClientCalls;

import static io.grpc.MethodDescriptor.generateFullMethodName;
import static io.grpc.stub.ClientCalls.blockingUnaryCall;
import static io.grpc.stub.ClientCalls.futureUnaryCall;
import static io.grpc.stub.ServerCalls.asyncUnaryCall;
import static io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall;

/**
 * <pre>
 * service
 * </pre>
 */
@javax.annotation.Generated(
    value = "by gRPC proto compiler (version 1.11.0)",
    comments = "Source: ReportService.proto")
public final class ReportServiceGrpc {

  private ReportServiceGrpc() {}

  public static final String SERVICE_NAME = "org.wso2.apim.grpc.telemetry.receiver.generated.ReportService";

  // Static method descriptors that strictly reflect the proto.
  @io.grpc.ExperimentalApi("https://github.com/grpc/grpc-java/issues/1901")
  @Deprecated // Use {@link #getReportMethod()} instead.
  public static final io.grpc.MethodDescriptor<org.wso2.apim.grpc.telemetry.receiver.generated.ReportRequest,
      org.wso2.apim.grpc.telemetry.receiver.generated.ReportResponse> METHOD_REPORT = getReportMethodHelper();

  private static volatile io.grpc.MethodDescriptor<org.wso2.apim.grpc.telemetry.receiver.generated.ReportRequest,
      org.wso2.apim.grpc.telemetry.receiver.generated.ReportResponse> getReportMethod;

  @io.grpc.ExperimentalApi("https://github.com/grpc/grpc-java/issues/1901")
  public static io.grpc.MethodDescriptor<org.wso2.apim.grpc.telemetry.receiver.generated.ReportRequest,
      org.wso2.apim.grpc.telemetry.receiver.generated.ReportResponse> getReportMethod() {
    return getReportMethodHelper();
  }

  private static io.grpc.MethodDescriptor<org.wso2.apim.grpc.telemetry.receiver.generated.ReportRequest,
      org.wso2.apim.grpc.telemetry.receiver.generated.ReportResponse> getReportMethodHelper() {
    io.grpc.MethodDescriptor<org.wso2.apim.grpc.telemetry.receiver.generated.ReportRequest, org.wso2.apim.grpc.telemetry.receiver.generated.ReportResponse> getReportMethod;
    if ((getReportMethod = ReportServiceGrpc.getReportMethod) == null) {
      synchronized (ReportServiceGrpc.class) {
        if ((getReportMethod = ReportServiceGrpc.getReportMethod) == null) {
          ReportServiceGrpc.getReportMethod = getReportMethod =
              io.grpc.MethodDescriptor.<org.wso2.apim.grpc.telemetry.receiver.generated.ReportRequest, org.wso2.apim.grpc.telemetry.receiver.generated.ReportResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(
                  "org.wso2.apim.grpc.telemetry.receiver.generated.ReportService", "report"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  org.wso2.apim.grpc.telemetry.receiver.generated.ReportRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  org.wso2.apim.grpc.telemetry.receiver.generated.ReportResponse.getDefaultInstance()))
                  .setSchemaDescriptor(new ReportServiceMethodDescriptorSupplier("report"))
                  .build();
          }
        }
     }
     return getReportMethod;
  }

  /**
   * Creates a new async stub that supports all call types for the service
   */
  public static ReportServiceStub newStub(io.grpc.Channel channel) {
    return new ReportServiceStub(channel);
  }

  /**
   * Creates a new blocking-style stub that supports unary and streaming output calls on the service
   */
  public static ReportServiceBlockingStub newBlockingStub(
      io.grpc.Channel channel) {
    return new ReportServiceBlockingStub(channel);
  }

  /**
   * Creates a new ListenableFuture-style stub that supports unary calls on the service
   */
  public static ReportServiceFutureStub newFutureStub(
      io.grpc.Channel channel) {
    return new ReportServiceFutureStub(channel);
  }

  /**
   * <pre>
   * service
   * </pre>
   */
  public static abstract class ReportServiceImplBase implements io.grpc.BindableService {

    /**
     */
    public void report(org.wso2.apim.grpc.telemetry.receiver.generated.ReportRequest request,
        io.grpc.stub.StreamObserver<org.wso2.apim.grpc.telemetry.receiver.generated.ReportResponse> responseObserver) {
      asyncUnimplementedUnaryCall(getReportMethodHelper(), responseObserver);
    }

    @Override public final io.grpc.ServerServiceDefinition bindService() {
      return io.grpc.ServerServiceDefinition.builder(getServiceDescriptor())
          .addMethod(
            getReportMethodHelper(),
            asyncUnaryCall(
              new MethodHandlers<
                org.wso2.apim.grpc.telemetry.receiver.generated.ReportRequest,
                org.wso2.apim.grpc.telemetry.receiver.generated.ReportResponse>(
                  this, METHODID_REPORT)))
          .build();
    }
  }

  /**
   * <pre>
   * service
   * </pre>
   */
  public static final class ReportServiceStub extends io.grpc.stub.AbstractStub<ReportServiceStub> {
    private ReportServiceStub(io.grpc.Channel channel) {
      super(channel);
    }

    private ReportServiceStub(io.grpc.Channel channel,
        io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @Override
    protected ReportServiceStub build(io.grpc.Channel channel,
        io.grpc.CallOptions callOptions) {
      return new ReportServiceStub(channel, callOptions);
    }

    /**
     */
    public void report(org.wso2.apim.grpc.telemetry.receiver.generated.ReportRequest request,
        io.grpc.stub.StreamObserver<org.wso2.apim.grpc.telemetry.receiver.generated.ReportResponse> responseObserver) {
      ClientCalls.asyncUnaryCall(
          getChannel().newCall(getReportMethodHelper(), getCallOptions()), request, responseObserver);
    }
  }

  /**
   * <pre>
   * service
   * </pre>
   */
  public static final class ReportServiceBlockingStub extends io.grpc.stub.AbstractStub<ReportServiceBlockingStub> {
    private ReportServiceBlockingStub(io.grpc.Channel channel) {
      super(channel);
    }

    private ReportServiceBlockingStub(io.grpc.Channel channel,
        io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @Override
    protected ReportServiceBlockingStub build(io.grpc.Channel channel,
        io.grpc.CallOptions callOptions) {
      return new ReportServiceBlockingStub(channel, callOptions);
    }

    /**
     */
    public org.wso2.apim.grpc.telemetry.receiver.generated.ReportResponse report(org.wso2.apim.grpc.telemetry.receiver.generated.ReportRequest request) {
      return blockingUnaryCall(
          getChannel(), getReportMethodHelper(), getCallOptions(), request);
    }
  }

  /**
   * <pre>
   * service
   * </pre>
   */
  public static final class ReportServiceFutureStub extends io.grpc.stub.AbstractStub<ReportServiceFutureStub> {
    private ReportServiceFutureStub(io.grpc.Channel channel) {
      super(channel);
    }

    private ReportServiceFutureStub(io.grpc.Channel channel,
        io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @Override
    protected ReportServiceFutureStub build(io.grpc.Channel channel,
        io.grpc.CallOptions callOptions) {
      return new ReportServiceFutureStub(channel, callOptions);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<org.wso2.apim.grpc.telemetry.receiver.generated.ReportResponse> report(
        org.wso2.apim.grpc.telemetry.receiver.generated.ReportRequest request) {
      return futureUnaryCall(
          getChannel().newCall(getReportMethodHelper(), getCallOptions()), request);
    }
  }

  private static final int METHODID_REPORT = 0;

  private static final class MethodHandlers<Req, Resp> implements
      io.grpc.stub.ServerCalls.UnaryMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.ServerStreamingMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.ClientStreamingMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.BidiStreamingMethod<Req, Resp> {
    private final ReportServiceImplBase serviceImpl;
    private final int methodId;

    MethodHandlers(ReportServiceImplBase serviceImpl, int methodId) {
      this.serviceImpl = serviceImpl;
      this.methodId = methodId;
    }

    @Override
    @SuppressWarnings("unchecked")
    public void invoke(Req request, io.grpc.stub.StreamObserver<Resp> responseObserver) {
      switch (methodId) {
        case METHODID_REPORT:
          serviceImpl.report((org.wso2.apim.grpc.telemetry.receiver.generated.ReportRequest) request,
              (io.grpc.stub.StreamObserver<org.wso2.apim.grpc.telemetry.receiver.generated.ReportResponse>) responseObserver);
          break;
        default:
          throw new AssertionError();
      }
    }

    @Override
    @SuppressWarnings("unchecked")
    public io.grpc.stub.StreamObserver<Req> invoke(
        io.grpc.stub.StreamObserver<Resp> responseObserver) {
      switch (methodId) {
        default:
          throw new AssertionError();
      }
    }
  }

  private static abstract class ReportServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoFileDescriptorSupplier, io.grpc.protobuf.ProtoServiceDescriptorSupplier {
    ReportServiceBaseDescriptorSupplier() {}

    @Override
    public com.google.protobuf.Descriptors.FileDescriptor getFileDescriptor() {
      return org.wso2.apim.grpc.telemetry.receiver.generated.ReportServiceOuterClass.getDescriptor();
    }

    @Override
    public com.google.protobuf.Descriptors.ServiceDescriptor getServiceDescriptor() {
      return getFileDescriptor().findServiceByName("ReportService");
    }
  }

  private static final class ReportServiceFileDescriptorSupplier
      extends ReportServiceBaseDescriptorSupplier {
    ReportServiceFileDescriptorSupplier() {}
  }

  private static final class ReportServiceMethodDescriptorSupplier
      extends ReportServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoMethodDescriptorSupplier {
    private final String methodName;

    ReportServiceMethodDescriptorSupplier(String methodName) {
      this.methodName = methodName;
    }

    @Override
    public com.google.protobuf.Descriptors.MethodDescriptor getMethodDescriptor() {
      return getServiceDescriptor().findMethodByName(methodName);
    }
  }

  private static volatile io.grpc.ServiceDescriptor serviceDescriptor;

  public static io.grpc.ServiceDescriptor getServiceDescriptor() {
    io.grpc.ServiceDescriptor result = serviceDescriptor;
    if (result == null) {
      synchronized (ReportServiceGrpc.class) {
        result = serviceDescriptor;
        if (result == null) {
          serviceDescriptor = result = io.grpc.ServiceDescriptor.newBuilder(SERVICE_NAME)
              .setSchemaDescriptor(new ReportServiceFileDescriptorSupplier())
              .addMethod(getReportMethodHelper())
              .build();
        }
      }
    }
    return result;
  }
}
